package application

import (
	"context"
	"event-sourcing-service-test/internal/domain"
	"event-sourcing-service-test/internal/eventstore"
)

type Queries struct{ Store eventstore.Store }

func (q Queries) WaitForProjection(c context.Context) error {
	return eventstore.WaitForProjection(c)
}

func (q Queries) Events(c context.Context, t, typ, id string) ([]domain.Event, error) {
	return q.Store.Load(c, t, typ, id)
}
func (q Queries) Scan(c context.Context, t string, pos int64, n int) (eventstore.Page, error) {
	return q.Store.Scan(c, t, pos, n)
}
