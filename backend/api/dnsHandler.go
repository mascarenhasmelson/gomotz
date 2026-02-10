package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/mascarenhasmelson/gomotz/bgservices"
	"github.com/mascarenhasmelson/gomotz/utils"
)

func (r *Router) handleDNSPost(w http.ResponseWriter, req *http.Request) {
	var request utils.Request

	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON request body", http.StatusBadRequest)
		return
	}

	r.processDNS(w, request)
}
func (r *Router) handleDNSGet(w http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()

	request := utils.Request{
		Domain: q.Get("domain"),
		Type:   q.Get("type"),
		Server: q.Get("server"),
	}

	r.processDNS(w, request)
}
func (r *Router) processDNS(w http.ResponseWriter, req utils.Request) {
	if req.Domain == "" {
		http.Error(w, `{"error":"domain is required"}`, http.StatusBadRequest)
		return
	}

	if req.Type == "" {
		req.Type = "A"
	}

	start := time.Now()
	answers, err := bgservices.QueryRR(req)
	duration := time.Since(start)

	resp := utils.APIResponse{
		Domain:   req.Domain,
		Type:     strings.ToUpper(req.Type),
		Server:   req.Server,
		Answers:  answers,
		Duration: duration.String(),
	}

	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		resp.Status = "error"
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":  err.Error(),
			"domain": req.Domain,
			"type":   resp.Type,
		})
		return
	}

	resp.Status = "success"
	json.NewEncoder(w).Encode(resp)
}
