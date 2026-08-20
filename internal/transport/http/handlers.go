package httpapi

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"event-sourcing-service/internal/application"
	"event-sourcing-service/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
	"strconv"
	"strings"
)

type API struct {
	Commands   application.Commands
	Queries    application.Queries
	DB         *pgxpool.Pool
	Replay     application.Replay
	Check      application.Consistency
	AdminToken string
}

func (a API) ReplayCreate(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Tenant-ID") == "" {
		writeError(w, r, 400, "INVALID_ARGUMENT", "X-Tenant-ID is required")
		return
	}
	if a.AdminToken == "" || subtle.ConstantTimeCompare([]byte(a.AdminToken), []byte(r.Header.Get("X-Admin-Token"))) != 1 {
		writeError(w, r, 403, "FORBIDDEN", "admin token required")
		return
	}
	var q struct {
		Projection string `json:"projection"`
	}
	if json.NewDecoder(r.Body).Decode(&q) != nil || q.Projection == "" {
		writeError(w, r, 400, "INVALID_ARGUMENT", "projection is required")
		return
	}
	id, e := a.Replay.Start(r.Context(), r.Header.Get("X-Tenant-ID"), q.Projection)
	if e != nil {
		writeDomainError(w, r, e)
		return
	}
	writeJSON(w, 202, map[string]any{"data": map[string]any{"job_id": id, "status": "queued"}})
}
func (a API) ReplayGet(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Tenant-ID") == "" {
		writeError(w, r, 400, "INVALID_ARGUMENT", "X-Tenant-ID is required")
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/v1/replays/")
	v, e := a.Replay.Jobs.Get(r.Context(), r.Header.Get("X-Tenant-ID"), id)
	if e != nil {
		writeError(w, r, 404, "NOT_FOUND", "replay job not found")
		return
	}
	writeJSON(w, 200, map[string]any{"data": v})
}
func (a API) CheckHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Tenant-ID") == "" {
		writeError(w, r, 400, "INVALID_ARGUMENT", "X-Tenant-ID is required")
		return
	}
	var q struct {
		Type string `json:"aggregate_type"`
		ID   string `json:"aggregate_id"`
	}
	if json.NewDecoder(r.Body).Decode(&q) != nil || q.Type == "" || q.ID == "" {
		writeError(w, r, 400, "INVALID_ARGUMENT", "aggregate_type and aggregate_id are required")
		return
	}
	status, e := a.Check.Check(r.Context(), r.Header.Get("X-Tenant-ID"), q.Type, q.ID)
	if e != nil {
		writeDomainError(w, r, e)
		return
	}
	writeJSON(w, 200, map[string]any{"data": map[string]any{"status": status}})
}

func (a API) Ready(w http.ResponseWriter, r *http.Request) {
	if a.DB == nil || a.DB.Ping(r.Context()) != nil {
		writeError(w, r, 503, "NOT_READY", "database is unavailable")
		return
	}
	writeJSON(w, 200, map[string]any{"status": "ready"})
}
func (a API) View(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		writeError(w, r, 404, "NOT_FOUND", "invalid view path")
		return
	}
	tenant := r.Header.Get("X-Tenant-ID")
	if tenant == "" {
		writeError(w, r, 400, "INVALID_ARGUMENT", "X-Tenant-ID is required")
		return
	}
	var data map[string]any
	var err error
	if parts[2] == "orders" {
		var buyer, currency, status string
		var amount, version int64
		err = a.DB.QueryRow(r.Context(), `select buyer_id,currency,amount,status,version from order_views where tenant_id=$1 and order_id=$2`, tenant, parts[3]).Scan(&buyer, &currency, &amount, &status, &version)
		data = map[string]any{"order_id": parts[3], "buyer_id": buyer, "currency": currency, "amount": amount, "status": status, "version": version}
	} else if parts[2] == "accounts" {
		var currency string
		var balance, credited, debited, version int64
		err = a.DB.QueryRow(r.Context(), `select currency,balance,credited,debited,version from account_views where tenant_id=$1 and account_id=$2`, tenant, parts[3]).Scan(&currency, &balance, &credited, &debited, &version)
		data = map[string]any{"account_id": parts[3], "currency": currency, "balance": balance, "credited": credited, "debited": debited, "version": version}
	} else {
		writeError(w, r, 404, "NOT_FOUND", "unknown view")
		return
	}
	if err != nil {
		writeError(w, r, 404, "NOT_FOUND", "view not found")
		return
	}
	writeJSON(w, 200, map[string]any{"data": data})
}

func (a API) CommandsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Tenant-ID") == "" {
		writeError(w, r, 400, "INVALID_ARGUMENT", "X-Tenant-ID is required")
		return
	}
	var q struct {
		CommandType string         `json:"command_type"`
		Expected    int64          `json:"expected_version"`
		Key         string         `json:"idempotency_key"`
		Params      map[string]any `json:"params"`
	}
	if json.NewDecoder(r.Body).Decode(&q) != nil {
		writeError(w, r, 400, "INVALID_ARGUMENT", "invalid json")
		return
	}
	x := strings.Split(r.URL.Path, "/")
	if len(x) < 6 {
		writeError(w, r, 404, "NOT_FOUND", "invalid path")
		return
	}
	ev, e := a.Commands.Execute(r.Context(), r.Header.Get("X-Tenant-ID"), x[3], x[4], q.CommandType, q.Key, q.Expected, q.Params)
	if e != nil {
		writeDomainError(w, r, e)
		return
	}
	writeJSON(w, 201, map[string]any{"request_id": r.Header.Get("X-Request-ID"), "data": map[string]any{"events": ev}})
}
func (a API) Events(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Tenant-ID") == "" {
		writeError(w, r, 400, "INVALID_ARGUMENT", "X-Tenant-ID is required")
		return
	}
	x := strings.Split(r.URL.Path, "/")
	if len(x) < 6 {
		writeError(w, r, 404, "NOT_FOUND", "invalid path")
		return
	}
	ev, e := a.Queries.Events(r.Context(), r.Header.Get("X-Tenant-ID"), x[3], x[4])
	if e != nil {
		writeDomainError(w, r, e)
		return
	}
	writeJSON(w, 200, map[string]any{"request_id": r.Header.Get("X-Request-ID"), "data": map[string]any{"items": ev}})
}

func (a API) GlobalEvents(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Tenant-ID") == "" {
		writeError(w, r, 400, "INVALID_ARGUMENT", "X-Tenant-ID is required")
		return
	}
	pos := int64(0)
	if raw := r.URL.Query().Get("from_position"); raw != "" {
		v, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || v < 0 {
			writeError(w, r, 400, "INVALID_ARGUMENT", "invalid from_position")
			return
		}
		pos = v
	}
	n := 100
	if raw := r.URL.Query().Get("limit"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 || v > 500 {
			writeError(w, r, 400, "INVALID_ARGUMENT", "limit must be between 1 and 500")
			return
		}
		n = v
	}
	page, e := a.Queries.Scan(r.Context(), r.Header.Get("X-Tenant-ID"), pos, n)
	if e != nil {
		writeDomainError(w, r, e)
		return
	}
	writeJSON(w, 200, map[string]any{"request_id": r.Header.Get("X-Request-ID"), "data": map[string]any{"items": page.Events, "next_position": page.Next}})
}

func writeJSON(w http.ResponseWriter, s int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(s)
	_ = json.NewEncoder(w).Encode(v)
}
func writeError(w http.ResponseWriter, r *http.Request, s int, c, m string) {
	writeJSON(w, s, map[string]any{"request_id": r.Header.Get("X-Request-ID"), "error": map[string]string{"code": c, "message": m}})
}
func writeDomainError(w http.ResponseWriter, r *http.Request, e error) {
	s, c := 500, "INTERNAL_ERROR"
	switch {
	case errors.Is(e, domain.ErrInvalid):
		s, c = 400, "INVALID_ARGUMENT"
	case errors.Is(e, domain.ErrConflict):
		s, c = 409, "CONFLICT"
	case errors.Is(e, domain.ErrInsufficient):
		s, c = 409, "INSUFFICIENT_FUNDS"
	case errors.Is(e, domain.ErrNotFound):
		s, c = 404, "NOT_FOUND"
	}
	writeError(w, r, s, c, e.Error())
}
