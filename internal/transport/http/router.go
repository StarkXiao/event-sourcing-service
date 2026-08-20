package httpapi

import (
	"event-sourcing-service/internal/application"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
)

func Router(a application.Commands, q application.Queries, db *pgxpool.Pool, replay application.Replay, check application.Consistency, adminToken string) http.Handler {
	api := API{Commands: a, Queries: q, DB: db, Replay: replay, Check: check, AdminToken: adminToken}
	m := http.NewServeMux()
	m.HandleFunc("/v1/events", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.Header().Set("Allow", "GET")
			writeError(w, r, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method is not allowed")
			return
		}
		api.GlobalEvents(w, r)
	})
	m.HandleFunc("/v1/views/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			writeError(w, r, 405, "METHOD_NOT_ALLOWED", "method is not allowed")
			return
		}
		api.View(w, r)
	})
	m.HandleFunc("/v1/projections/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			writeError(w, r, 405, "METHOD_NOT_ALLOWED", "method is not allowed")
			return
		}
		api.ReplayCreate(w, r)
	})
	m.HandleFunc("/v1/replays/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			writeError(w, r, 405, "METHOD_NOT_ALLOWED", "method is not allowed")
			return
		}
		api.ReplayGet(w, r)
	})
	m.HandleFunc("/v1/consistency/checks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			writeError(w, r, 405, "METHOD_NOT_ALLOWED", "method is not allowed")
			return
		}
		api.CheckHandler(w, r)
	})
	m.HandleFunc("/v1/aggregates/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			api.CommandsHandler(w, r)
		case "GET":
			api.Events(w, r)
		default:
			w.Header().Set("Allow", "GET, POST")
			writeError(w, r, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method is not allowed")
		}
	})
	m.HandleFunc("/health/live", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200); w.Write([]byte(`{"status":"ok"}`)) })
	m.HandleFunc("/health/ready", api.Ready)
	return RequestID(m)
}
