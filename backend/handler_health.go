package main

import (
	"encoding/json"
	"net/http"
)

type adapterStatus struct {
	Available bool   `json:"available"`
	Reason    string `json:"reason,omitempty"`
}

type healthResponse struct {
	Status   string                   `json:"status"`
	Adapters map[string]adapterStatus `json:"adapters"`
}

func handleHealth(adapters map[string]LLMAdapter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		statuses := make(map[string]adapterStatus, len(adapters))
		for id, a := range adapters {
			s := adapterStatus{Available: a.Available()}
			if !s.Available {
				s.Reason = unavailableReason(a)
			}
			statuses[id] = s
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(healthResponse{
			Status:   "ok",
			Adapters: statuses,
		})
	}
}

// unavailableReason returns a human-readable reason why an adapter is unavailable.
func unavailableReason(a LLMAdapter) string {
	switch a.(type) {
	case *ClaudeAdapter:
		return "no API key"
	case *OllamaAdapter:
		return "ollama unreachable"
	default:
		return "unavailable"
	}
}
