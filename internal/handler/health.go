package handler

import (
	"encoding/json"
	"net/http"

	"github.com/mlorentedev/pollex/internal/adapter"
	"github.com/mlorentedev/pollex/internal/metrics"
)

type adapterStatus struct {
	Available bool   `json:"available"`
	Reason    string `json:"reason,omitempty"`
}

type healthResponse struct {
	Status   string                   `json:"status"`
	Adapters map[string]adapterStatus `json:"adapters"`
}

func Health(adapters map[string]adapter.LLMAdapter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		statuses := make(map[string]adapterStatus, len(adapters))
		for id, a := range adapters {
			s := adapterStatus{Available: a.Available()}
			if s.Available {
				metrics.AdapterAvailable.WithLabelValues(id).Set(1)
			} else {
				metrics.AdapterAvailable.WithLabelValues(id).Set(0)
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

func unavailableReason(a adapter.LLMAdapter) string {
	switch a.(type) {
	case *adapter.ClaudeAdapter:
		return "no API key"
	case *adapter.OllamaAdapter:
		return "ollama unreachable"
	case *adapter.LlamaCppAdapter:
		return "llama-server unreachable"
	default:
		return "unavailable"
	}
}
