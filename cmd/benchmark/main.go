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

type modelsResponse struct {
	Models []struct {
		ID       string `json:"id"`
		Provider string `json:"provider"`
	} `json:"models"`
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
}

func main() {
	url := flag.String("url", "http://localhost:8090", "API base URL")
	apiKey := flag.String("api-key", "", "API key (optional)")
	runs := flag.Int("runs", 3, "Number of runs per sample")
	model := flag.String("model", "", "Model ID to use (default: first available)")
	flag.Parse()

	baseURL := strings.TrimRight(*url, "/")
	client := &http.Client{Timeout: 120 * time.Second}

	// Discover models
	modelID := *model
	if modelID == "" {
		modelID = discoverModel(client, baseURL, *apiKey)
	}
	fmt.Printf("Benchmarking against %s using model: %s (%d runs per sample)\n\n", baseURL, modelID, *runs)

	// Run benchmarks
	var results []result
	for _, sample := range Samples {
		for run := 1; run <= *runs; run++ {
			fmt.Printf("  Running %s (run %d/%d)...", sample.Name, run, *runs)
			r := benchmark(client, baseURL, *apiKey, modelID, sample, run)
			results = append(results, r)
			fmt.Printf(" %dms\n", r.ElapsedMs)
		}
	}

	// Print results table
	fmt.Println()
	printTable(results)
	printSummary(results)
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

	var models modelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&models); err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding models: %v\n", err)
		os.Exit(1)
	}

	if len(models.Models) == 0 {
		fmt.Fprintln(os.Stderr, "No models available")
		os.Exit(1)
	}

	return models.Models[0].ID
}

func benchmark(client *http.Client, baseURL, apiKey, modelID string, sample Sample, run int) result {
	payload, _ := json.Marshal(polishRequest{
		Text:    sample.Text,
		ModelID: modelID,
	})

	req, err := http.NewRequest("POST", baseURL+"/api/polish", strings.NewReader(string(payload)))
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nError creating request: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("X-API-Key", apiKey)
	}

	start := time.Now()
	resp, err := client.Do(req)
	wallMs := time.Since(start).Milliseconds()

	if err != nil {
		fmt.Fprintf(os.Stderr, "\nError sending request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Fprintf(os.Stderr, "\nPolish returned %d: %s\n", resp.StatusCode, body)
		os.Exit(1)
	}

	var pr polishResponse
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		fmt.Fprintf(os.Stderr, "\nError decoding response: %v\n", err)
		os.Exit(1)
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
		ratio := float64(r.OutChars) / float64(r.Chars)
		fmt.Printf("| %-6s | %5d | %-20s | %d | %12d | %9d | %9d | %5.2f |\n",
			r.Sample, r.Chars, r.Model, r.Run, r.ElapsedMs, r.WallMs, r.OutChars, ratio)
	}
}

func printSummary(results []result) {
	if len(results) == 0 {
		return
	}

	var totalElapsed int64
	var totalChars int
	minElapsed := results[0].ElapsedMs
	maxElapsed := results[0].ElapsedMs
	minSample := results[0].Sample
	maxSample := results[0].Sample

	for _, r := range results {
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
	fmt.Printf("- Total runs: %d\n", len(results))
}
