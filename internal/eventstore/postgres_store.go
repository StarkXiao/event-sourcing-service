package eventstore

import (
	"context"
	"encoding/json"
	"errors"
	"event-sourcing-service/internal/domain"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type Postgres struct{ DB *pgxpool.Pool }

func (p Postgres) Load(c context.Context, t, typ, id string) (out []domain.Event, e error) {
	rows, e := p.DB.Query(c, `select id,tenant_id,aggregate_type,aggregate_id,aggregate_version,event_type,payload,metadata,occurred_at,global_position,idempotency_key from events where tenant_id=$1 and aggregate_type=$2 and aggregate_id=$3 order by aggregate_version`, t, typ, id)
	if e != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var v domain.Event
		var a, b []byte
		e = rows.Scan(&v.ID, &v.TenantID, &v.AggregateType, &v.AggregateID, &v.Version, &v.Type, &a, &b, &v.OccurredAt, &v.Position, &v.IdempotencyKey)
		if e != nil {
			return
		}
		if e = json.Unmarshal(a, &v.Payload); e != nil {
			return
		}
		if e = json.Unmarshal(b, &v.Metadata); e != nil {
			return
		}
		out = append(out, v)
	}
	e = rows.Err()
	return
}
func (p Postgres) Append(c context.Context, r AppendRequest) ([]domain.Event, error) {
	tx, e := p.DB.Begin(c)
	if e != nil {
		return nil, e
	}
	defer tx.Rollback(c)
	var existingType, existingID string
	var firstVersion int64
	var eventCount int
	e = tx.QueryRow(c, "select aggregate_type,aggregate_id,first_version,event_count from idempotency_records where tenant_id=$1 and idempotency_key=$2", r.TenantID, r.IdempotencyKey).Scan(&existingType, &existingID, &firstVersion, &eventCount)
	if e == nil {
		if existingType != r.Type || existingID != r.ID {
			return nil, domain.ErrConflict
		}
		rows, qe := tx.Query(c, `select id,tenant_id,aggregate_type,aggregate_id,aggregate_version,event_type,payload,metadata,occurred_at,global_position,idempotency_key from events where tenant_id=$1 and aggregate_type=$2 and aggregate_id=$3 and aggregate_version >= $4 and aggregate_version < $4+$5 order by aggregate_version`, r.TenantID, existingType, existingID, firstVersion, eventCount)
		if qe != nil {
			return nil, qe
		}
		defer rows.Close()
		var result []domain.Event
		for rows.Next() {
			var v domain.Event
			var a, b []byte
			if qe = rows.Scan(&v.ID, &v.TenantID, &v.AggregateType, &v.AggregateID, &v.Version, &v.Type, &a, &b, &v.OccurredAt, &v.Position, &v.IdempotencyKey); qe != nil {
				return nil, qe
			}
			if qe = json.Unmarshal(a, &v.Payload); qe != nil {
				return nil, qe
			}
			if qe = json.Unmarshal(b, &v.Metadata); qe != nil {
				return nil, qe
			}
			result = append(result, v)
		}
		if qe = rows.Err(); qe != nil {
			return nil, qe
		}
		return result, nil
	}
	if !errors.Is(e, pgx.ErrNoRows) {
		return nil, e
	}
	var ver int64
	_, e = tx.Exec(c, `insert into aggregate_heads(tenant_id,aggregate_type,aggregate_id,version) values($1,$2,$3,0) on conflict do nothing`, r.TenantID, r.Type, r.ID)
	if e != nil {
		return nil, e
	}
	e = tx.QueryRow(c, "select version from aggregate_heads where tenant_id=$1 and aggregate_type=$2 and aggregate_id=$3 for update", r.TenantID, r.Type, r.ID).Scan(&ver)
	if errors.Is(e, pgx.ErrNoRows) {
		ver = 0
		e = nil
	}
	if e != nil {
		return nil, e
	}
	if ver != r.ExpectedVersion {
		return nil, domain.ErrConflict
	}
	for i := range r.Events {
		v := &r.Events[i]
		v.ID = uuid.NewString()
		v.TenantID = r.TenantID
		v.AggregateType = r.Type
		v.AggregateID = r.ID
		v.Version = ver + int64(i) + 1
		v.IdempotencyKey = fmt.Sprintf("%s:%s", r.IdempotencyKey, v.ID)
		v.OccurredAt = time.Now().UTC()
		if v.Metadata == nil {
			v.Metadata = map[string]any{}
		}
		a, marshalErr := json.Marshal(v.Payload)
		if marshalErr != nil {
			return nil, marshalErr
		}
		b, marshalErr := json.Marshal(v.Metadata)
		if marshalErr != nil {
			return nil, marshalErr
		}
		e = tx.QueryRow(c, `insert into events(id,tenant_id,aggregate_type,aggregate_id,aggregate_version,event_type,payload,metadata,occurred_at,idempotency_key) values($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) returning global_position`, v.ID, v.TenantID, v.AggregateType, v.AggregateID, v.Version, v.Type, a, b, v.OccurredAt, v.IdempotencyKey).Scan(&v.Position)
		if e != nil {
			return nil, e
		}
		for _, name := range projections(r.Type) {
			_, e = tx.Exec(c, "insert into outbox_tasks(tenant_id,event_id,projection) values($1,$2,$3)", r.TenantID, v.ID, name)
			if e != nil {
				return nil, e
			}
		}
	}
	_, e = tx.Exec(c, `insert into aggregate_heads(tenant_id,aggregate_type,aggregate_id,version) values($1,$2,$3,$4) on conflict(tenant_id,aggregate_type,aggregate_id) do update set version=excluded.version,updated_at=now()`, r.TenantID, r.Type, r.ID, ver+int64(len(r.Events)))
	if e != nil {
		return nil, e
	}
	_, e = tx.Exec(c, `insert into idempotency_records(tenant_id,idempotency_key,aggregate_type,aggregate_id,first_version,event_count) values($1,$2,$3,$4,$5,$6)`, r.TenantID, r.IdempotencyKey, r.Type, r.ID, ver+1, len(r.Events))
	if e != nil {
		return nil, e
	}
	return r.Events, tx.Commit(c)
}
func projections(t string) []string {
	if t == "order" {
		return []string{"orders"}
	}
	return []string{"accounts"}
}
func (p Postgres) Scan(c context.Context, t string, pos int64, n int) (Page, error) {
	rows, e := p.DB.Query(c, `select id from events where tenant_id=$1 and global_position>$2 order by global_position limit $3`, t, pos, n)
	if e != nil {
		return Page{}, e
	}
	defer rows.Close()
	x := Page{}
	for rows.Next() {
		var id string
		if e = rows.Scan(&id); e != nil {
			return x, e
		}
		v, e := p.Event(c, id)
		if e != nil {
			return x, e
		}
		x.Events = append(x.Events, v)
		x.Next = v.Position
	}
	return x, rows.Err()
}
func (p Postgres) Event(c context.Context, id string) (v domain.Event, e error) {
	var a, b []byte
	e = p.DB.QueryRow(c, `select id,tenant_id,aggregate_type,aggregate_id,aggregate_version,event_type,payload,metadata,occurred_at,global_position,idempotency_key from events where id=$1`, id).Scan(&v.ID, &v.TenantID, &v.AggregateType, &v.AggregateID, &v.Version, &v.Type, &a, &b, &v.OccurredAt, &v.Position, &v.IdempotencyKey)
	if e != nil {
		return
	}
	if e = json.Unmarshal(a, &v.Payload); e != nil {
		return
	}
	e = json.Unmarshal(b, &v.Metadata)
	return
}
