package api

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mascarenhasmelson/gomotz/servicetools"
	"github.com/mascarenhasmelson/gomotz/utils"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Router struct {
	ctx  context.Context
	pool *pgxpool.Pool
	mux  *http.ServeMux
}

func NewRouter(ctx context.Context, pool *pgxpool.Pool) *http.ServeMux {
	r := &Router{
		ctx:  ctx,
		pool: pool,
		mux:  http.NewServeMux(),
	}

	r.routes()
	return r.mux
}

var wsMu sync.Mutex

func (r *Router) routes() {
	r.mux.HandleFunc("/services", r.servicesHandler)
	r.mux.HandleFunc("/services/", r.serviceHandler)
	r.mux.HandleFunc("/v1/services/isp", r.ispHandler)
	r.mux.HandleFunc("/v1/scan", handleSynScan)
}
func handleSynScan(w http.ResponseWriter, r *http.Request) {
	ws, err := servicetools.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer ws.Close()
	var req utils.ScanRequest
	if err := ws.ReadJSON(&req); err != nil {
		return
	}
	go servicetools.SynScan(ws, req.Target)
	for {
		time.Sleep(5 * time.Second)
		wsMu.Lock()
		err := ws.WriteMessage(websocket.PingMessage, nil)
		wsMu.Unlock()
		if err != nil {
			return
		}
	}
}

func (r *Router) servicesHandler(w http.ResponseWriter, req *http.Request) {
	if EnableCORS(&w, req) {
		return
	}
	if req.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	switch req.Method {
	case http.MethodGet:
		HandleFetchServices(r.ctx, w, r.pool)
	case http.MethodPost:
		HandleAddService(r.ctx, w, r.pool, req)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (r *Router) serviceHandler(w http.ResponseWriter, req *http.Request) {
	if EnableCORS(&w, req) {
		return
	}

	idStr := strings.TrimPrefix(req.URL.Path, "/services/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid service ID", http.StatusBadRequest)
		return
	}

	if req.Method == http.MethodDelete {
		HandleDeleteService(r.ctx, w, r.pool, id)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (r *Router) ispHandler(w http.ResponseWriter, req *http.Request) {
	if EnableCORS(&w, req) {
		return
	}
	if req.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	switch req.Method {
	case http.MethodGet:
		HandleGetISPInfo(w, req)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
