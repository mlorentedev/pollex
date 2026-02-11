package main

import (
	"encoding/json"
	"net/http"
)

func handleModels(models []ModelInfo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models)
	}
}
