package replay

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Jobs struct{ DB *pgxpool.Pool }

func (j Jobs) Create(c context.Context, t, p string) (string, error) {
	id := uuid.NewString()
	_, e := j.DB.Exec(c, "insert into replay_jobs(id,tenant_id,projection,status) values($1,$2,$3,'queued')", id, t, p)
	return id, e
}
func (j Jobs) Get(c context.Context, tenant, id string) (map[string]any, error) {
	var status string
	var n int64
	e := j.DB.QueryRow(c, "select status,processed from replay_jobs where id=$1 and tenant_id=$2", id, tenant).Scan(&status, &n)
	return map[string]any{"id": id, "status": status, "processed": n}, e
}
