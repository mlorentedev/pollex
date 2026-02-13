package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/mlorentedev/pollex/internal/adapter"
)

const maxTextLength = 10000

type polishRequest struct {
	Text    string `json:"text"`
	ModelID string `json:"model_id"`
}

type polishResponse struct {
	Polished  string `json:"polished"`
	Model     string `json:"model"`
	ElapsedMs int64  `json:"elapsed_ms"`
}

func Polish(adapters map[string]adapter.LLMAdapter, systemPrompt string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		var req polishRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			var maxBytesErr *http.MaxBytesError
			if errors.As(err, &maxBytesErr) {
				writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
				return
			}
			writeError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}

		if req.Text == "" {
			writeError(w, http.StatusBadRequest, "text is required")
			return
		}
		if len(req.Text) > maxTextLength {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("text too long: %d characters (max %d)", len(req.Text), maxTextLength))
			return
		}
		if req.ModelID == "" {
			writeError(w, http.StatusBadRequest, "model_id is required")
			return
		}

		a, ok := adapters[req.ModelID]
		if !ok {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("unknown model: %s", req.ModelID))
			return
		}

		start := time.Now()
		polished, err := a.Polish(r.Context(), req.Text, systemPrompt)
		elapsed := time.Since(start)

		if err != nil {
			writeError(w, http.StatusBadGateway, fmt.Sprintf("polish failed: %v", err))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(polishResponse{
			Polished:  polished,
			Model:     req.ModelID,
			ElapsedMs: elapsed.Milliseconds(),
		})
	}
}
