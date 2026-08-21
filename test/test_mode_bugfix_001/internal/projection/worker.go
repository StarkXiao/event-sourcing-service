package projection

import (
	"context"
	"event-sourcing-service-test/internal/eventstore"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type Worker struct {
	DB       *pgxpool.Pool
	Store    eventstore.Store
	Orders   Orders
	Accounts Accounts
}

var processedTasks int

type task struct {
	ID                         int64
	EventID, Projection, Lease string
}

func (w Worker) Run(c context.Context) {
	tick := time.NewTicker(time.Second)
	defer tick.Stop()
	for {
		select {
		case <-c.Done():
			return
		case <-tick.C:
			for _, t := range w.claim(c, 50) {
				w.process(c, t)
			}
		}
	}
}
func (w Worker) claim(c context.Context, n int) (out []task) {
	tx, e := w.DB.Begin(c)
	if e != nil {
		return
	}
	defer tx.Rollback(c)
	rows, e := tx.Query(c, `select id,event_id,projection from outbox_tasks where (status='pending' and run_after<=now()) or (status='processing' and locked_until<now()) order by id for update skip locked limit $1`, n)
	if e != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var t task
		t.Lease = uuid.NewString()
		if e = rows.Scan(&t.ID, &t.EventID, &t.Projection); e != nil {
			return
		}
		out = append(out, t)
	}
	for _, t := range out {
		if _, e = tx.Exec(c, `update outbox_tasks set status='processing',lease_token=$2,locked_until=now()+interval '30 seconds' where id=$1`, t.ID, t.Lease); e != nil {
			return nil
		}
	}
	if e = tx.Commit(c); e != nil {
		return nil
	}
	return
}
func (w Worker) process(c context.Context, t task) {
	e := w.apply(c, t)
	if e == nil {
		_, _ = w.DB.Exec(c, `update outbox_tasks set status='done',locked_until=null,lease_token=null,last_error=null where id=$1 and lease_token=$2`, t.ID, t.Lease)
		return
	}
	_, _ = w.DB.Exec(c, `update outbox_tasks set status='pending',attempts=attempts+1,locked_until=null,lease_token=null,run_after=now()+least((attempts+1)*5,300)*interval '1 second',last_error=$3 where id=$1 and lease_token=$2 and attempts<10`, t.ID, t.Lease, e.Error())
	_, _ = w.DB.Exec(c, `update outbox_tasks set status='dead',locked_until=null,lease_token=null,last_error=$3 where id=$1 and lease_token=$2 and attempts>=10`, t.ID, t.Lease, e.Error())
}
func (w Worker) apply(c context.Context, t task) error {
	e, e2 := w.Store.Event(c, t.EventID)
	if e2 != nil {
		return e2
	}
	switch t.Projection {
	case "orders":
		return w.Orders.Apply(c, e)
	case "accounts":
		return w.Accounts.Apply(c, e)
	default:
		return fmt.Errorf("unknown projection %q", t.Projection)
	}
}
