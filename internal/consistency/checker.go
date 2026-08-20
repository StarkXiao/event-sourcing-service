package consistency

import (
	"context"
	"encoding/json"
	"event-sourcing-service/internal/domain"
	"event-sourcing-service/internal/eventstore"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Checker struct {
	Store eventstore.Store
	DB    *pgxpool.Pool
}

func (c Checker) Check(ctx context.Context, t, typ, id string) (string, error) {
	events, e := c.Store.Load(ctx, t, typ, id)
	if e != nil {
		return "error", e
	}
	if len(events) == 0 {
		return "not_found", domain.ErrNotFound
	}
	if c.DB == nil {
		return "unavailable", fmt.Errorf("projection database is unavailable")
	}
	status := "unsupported"
	expected := map[string]any{}
	actual := map[string]any{}
	if typ == "order" {
		var w domain.Order
		w.ID = id
		for _, v := range events {
			w.Apply(v)
		}
		expected = map[string]any{"version": w.Version, "status": w.Status}
		var s string
		var v int64
		e = c.DB.QueryRow(ctx, `select status,version from order_views where tenant_id=$1 and order_id=$2`, t, id).Scan(&s, &v)
		actual = map[string]any{"version": v, "status": s}
		if e != nil {
			status = "projection_missing"
		} else if v == w.Version && s == w.Status {
			status = "consistent"
		} else {
			status = "divergent"
		}
	} else if typ == "account" {
		var w domain.Account
		w.ID = id
		for _, v := range events {
			w.Apply(v)
		}
		expected = map[string]any{"version": w.Version, "balance": w.Balance}
		var b, v int64
		e = c.DB.QueryRow(ctx, `select balance,version from account_views where tenant_id=$1 and account_id=$2`, t, id).Scan(&b, &v)
		actual = map[string]any{"version": v, "balance": b}
		if e != nil {
			status = "projection_missing"
		} else if v == w.Version && b == w.Balance {
			status = "consistent"
		} else {
			status = "divergent"
		}
	} else {
		return status, domain.ErrInvalid
	}
	eb, _ := json.Marshal(expected)
	ab, _ := json.Marshal(actual)
	if _, e = c.DB.Exec(ctx, `insert into consistency_checks(id,tenant_id,aggregate_type,aggregate_id,status,expected,actual,detail) values($1,$2,$3,$4,$5,$6,$7,$8)`, uuid.NewString(), t, typ, id, status, eb, ab, status); e != nil {
		return "audit_failed", e
	}
	return status, nil
}
