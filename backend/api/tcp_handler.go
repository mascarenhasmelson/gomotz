package api

import (
	"encoding/json"
	"net/http"

	"github.com/mascarenhasmelson/gomotz/bgservices"
	"github.com/mascarenhasmelson/gomotz/utils"
)

func TcpCheckHandler(w http.ResponseWriter, r *http.Request) {
	if EnableCORS(&w, r) {
		return
	}
	if r.Method == http.MethodOptions {
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req utils.TCPCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, utils.TCPCheckResponse{
			Success: false,
			Status:  "error",
			Message: "Invalid JSON request",
		})
		return
	}
	if req.Host == "" || req.Port <= 0 || req.Port > 65535 {
		writeJSON(w, utils.TCPCheckResponse{
			Success: false,
			Status:  "error",
			Message: "Invalid host or port",
		})
		return
	}
	resp := bgservices.TcpCheck(req)
	writeJSON(w, resp)
}

func writeJSON(w http.ResponseWriter, resp utils.TCPCheckResponse) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
