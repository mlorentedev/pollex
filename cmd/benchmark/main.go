package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type modelInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Provider string `json:"provider"`
}

type polishRequest struct {
	Text    string `json:"text"`
	ModelID string `json:"model_id"`
}

type polishResponse struct {
	Polished  string `json:"polished"`
	Model     string `json:"model"`
	ElapsedMs int64  `json:"elapsed_ms"`
}

type result struct {
	Sample    string
	Chars     int
	Model     string
	Run       int
	ElapsedMs int64
	WallMs    int64
	OutChars  int
	Error     string
}

func main() {
	url := flag.String("url", "http://localhost:8090", "API base URL")
	apiKey := flag.String("api-key", "", "API key (optional)")
	runs := flag.Int("runs", 3, "Number of runs per sample")
	model := flag.String("model", "", "Model ID to use (default: first available)")
	quality := flag.Bool("quality", false, "Quality mode: show input/output for each sample (1 run, no timing table)")
	jsonOut := flag.String("json", "", "Write results to JSON file (e.g. results.json)")
	warmup := flag.Bool("warmup", false, "Run one warmup request per sample before measuring")
	flag.Parse()

	baseURL := strings.TrimRight(*url, "/")
	client := &http.Client{Timeout: 180 * time.Second}

	// Discover models
	modelID := *model
	if modelID == "" {
		modelID = discoverModel(client, baseURL, *apiKey)
	}

	if *quality {
		runQualityMode(client, baseURL, *apiKey, modelID)
		return
	}

	fmt.Printf("Benchmarking against %s using model: %s (%d runs per sample", baseURL, modelID, *runs)
	if *warmup {
		fmt.Print(", warmup enabled")
	}
	fmt.Println(")")

	// Run benchmarks
	var results []result
	var failures int
	for _, sample := range Samples {
		if *warmup {
			fmt.Printf("  Warming up %s...", sample.Name)
			w := benchmark(client, baseURL, *apiKey, modelID, sample, 0)
			if w.Error != "" {
				fmt.Printf(" FAILED (%s)\n", w.Error)
			} else {
				fmt.Printf(" %dms (discarded)\n", w.ElapsedMs)
			}
		}
		for run := 1; run <= *runs; run++ {
			fmt.Printf("  Running %s (run %d/%d)...", sample.Name, run, *runs)
			r := benchmark(client, baseURL, *apiKey, modelID, sample, run)
			results = append(results, r)
			if r.Error != "" {
				fmt.Printf(" FAILED (%s)\n", r.Error)
				failures++
			} else {
				fmt.Printf(" %dms\n", r.ElapsedMs)
			}
		}
	}

	// Print results table
	fmt.Println()
	printTable(results)
	printSummary(results)

	// Write JSON output
	if *jsonOut != "" {
		if err := writeJSON(*jsonOut, results, baseURL, modelID); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing JSON: %v\n", err)
		} else {
			fmt.Printf("\nResults written to %s\n", *jsonOut)
		}
	}

	if failures > 0 {
		os.Exit(1)
	}
}

func discoverModel(client *http.Client, baseURL, apiKey string) string {
	req, err := http.NewRequest("GET", baseURL+"/api/models", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating request: %v\n", err)
		os.Exit(1)
	}
	if apiKey != "" {
		req.Header.Set("X-API-Key", apiKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching models: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Fprintf(os.Stderr, "Models endpoint returned %d: %s\n", resp.StatusCode, body)
		os.Exit(1)
	}

	var models []modelInfo
	if err := json.NewDecoder(resp.Body).Decode(&models); err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding models: %v\n", err)
		os.Exit(1)
	}

	if len(models) == 0 {
		fmt.Fprintln(os.Stderr, "No models available")
		os.Exit(1)
	}

	return models[0].ID
}

func benchmark(client *http.Client, baseURL, apiKey, modelID string, sample Sample, run int) result {
	fail := func(err string) result {
		return result{Sample: sample.Name, Chars: len(sample.Text), Run: run, Error: err}
	}

	payload, _ := json.Marshal(polishRequest{
		Text:    sample.Text,
		ModelID: modelID,
	})

	req, err := http.NewRequest("POST", baseURL+"/api/polish", strings.NewReader(string(payload)))
	if err != nil {
		return fail(err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("X-API-Key", apiKey)
	}

	start := time.Now()
	resp, err := client.Do(req)
	wallMs := time.Since(start).Milliseconds()

	if err != nil {
		return fail(err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fail(fmt.Sprintf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body))))
	}

	var pr polishResponse
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return fail(err.Error())
	}

	return result{
		Sample:    sample.Name,
		Chars:     len(sample.Text),
		Model:     pr.Model,
		Run:       run,
		ElapsedMs: pr.ElapsedMs,
		WallMs:    wallMs,
		OutChars:  len(pr.Polished),
	}
}

func printTable(results []result) {
	fmt.Println("| Sample | Chars | Model | Run | Elapsed (ms) | Wall (ms) | Out Chars | Ratio |")
	fmt.Println("|--------|-------|-------|-----|--------------|-----------|-----------|-------|")
	for _, r := range results {
		if r.Error != "" {
			fmt.Printf("| %-6s | %5d | %-20s | %d | %12s | %9s | %9s | %5s |\n",
				r.Sample, r.Chars, "-", r.Run, "FAIL", "-", "-", "-")
			continue
		}
		ratio := float64(r.OutChars) / float64(r.Chars)
		fmt.Printf("| %-6s | %5d | %-20s | %d | %12d | %9d | %9d | %5.2f |\n",
			r.Sample, r.Chars, r.Model, r.Run, r.ElapsedMs, r.WallMs, r.OutChars, ratio)
	}
}

func runQualityMode(client *http.Client, baseURL, apiKey, modelID string) {
	fmt.Printf("Quality test against %s using model: %s\n", baseURL, modelID)
	fmt.Println(strings.Repeat("=", 72))

	var failures int
	for i, sample := range QualitySamples {
		fmt.Printf("\n--- %d/%d: %s (%d chars) ---\n", i+1, len(QualitySamples), sample.Name, len(sample.Text))
		fmt.Printf("IN:  %s\n", sample.Text)

		payload, _ := json.Marshal(polishRequest{Text: sample.Text, ModelID: modelID})
		req, err := http.NewRequest("POST", baseURL+"/api/polish", strings.NewReader(string(payload)))
		if err != nil {
			fmt.Printf("ERR: %s\n", err)
			failures++
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		if apiKey != "" {
			req.Header.Set("X-API-Key", apiKey)
		}

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("ERR: %s\n", err)
			failures++
			continue
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			fmt.Printf("ERR: HTTP %d: %s\n", resp.StatusCode, strings.TrimSpace(string(body)))
			failures++
			continue
		}

		var pr polishResponse
		if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
			resp.Body.Close()
			fmt.Printf("ERR: %s\n", err)
			failures++
			continue
		}
		resp.Body.Close()

		fmt.Printf("OUT: %s\n", pr.Polished)
		fmt.Printf("     [%dms, %d->%d chars]\n", pr.ElapsedMs, len(sample.Text), len(pr.Polished))
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 72))
	fmt.Printf("Done: %d/%d passed\n", len(QualitySamples)-failures, len(QualitySamples))
	if failures > 0 {
		os.Exit(1)
	}
}

func printSummary(results []result) {
	var ok []result
	for _, r := range results {
		if r.Error == "" {
			ok = append(ok, r)
		}
	}

	failed := len(results) - len(ok)

	if len(ok) == 0 {
		fmt.Printf("\nSummary: all %d runs failed\n", len(results))
		return
	}

	var totalElapsed int64
	var totalChars int
	minElapsed := ok[0].ElapsedMs
	maxElapsed := ok[0].ElapsedMs
	minSample := ok[0].Sample
	maxSample := ok[0].Sample

	for _, r := range ok {
		totalElapsed += r.ElapsedMs
		totalChars += r.Chars
		if r.ElapsedMs < minElapsed {
			minElapsed = r.ElapsedMs
			minSample = r.Sample
		}
		if r.ElapsedMs > maxElapsed {
			maxElapsed = r.ElapsedMs
			maxSample = r.Sample
		}
	}

	avgMsPerChar := float64(totalElapsed) / float64(totalChars)

	fmt.Printf("\nSummary:\n")
	fmt.Printf("- Avg ms/char: %.2f\n", avgMsPerChar)
	fmt.Printf("- Min elapsed: %dms (%s)\n", minElapsed, minSample)
	fmt.Printf("- Max elapsed: %dms (%s)\n", maxElapsed, maxSample)
	fmt.Printf("- Total runs: %d (%d ok, %d failed)\n", len(results), len(ok), failed)
}

type jsonReport struct {
	Timestamp string   `json:"timestamp"`
	URL       string   `json:"url"`
	Model     string   `json:"model"`
	Results   []result `json:"results"`
}

func writeJSON(path string, results []result, baseURL, modelID string) error {
	report := jsonReport{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		URL:       baseURL,
		Model:     modelID,
		Results:   results,
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
