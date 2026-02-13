package handler

import (
	"encoding/json"
	"net/http"

	"github.com/mlorentedev/pollex/internal/adapter"
)

func Models(models []adapter.ModelInfo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models)
	}
}
