package application

import (
	"context"
	"event-sourcing-service/internal/domain"
	"event-sourcing-service/internal/eventstore"
)

type Queries struct{ Store eventstore.Store }

func (q Queries) Events(c context.Context, t, typ, id string) ([]domain.Event, error) {
	events, err := q.Store.Load(c, t, typ, id)
	if err != nil {
		return nil, err
	}
	return eventstore.CopyEvents(events), nil
}
func (q Queries) Scan(c context.Context, t string, pos int64, n int) (eventstore.Page, error) {
	return q.Store.Scan(c, t, pos, n)
}
