package application

import (
	"context"
	"event-sourcing-service/internal/domain"
	"event-sourcing-service/internal/eventstore"
	"testing"
)

type eventQueryStore struct {
	events []domain.Event
}

func (s eventQueryStore) Load(context.Context, string, string, string) ([]domain.Event, error) {
	return s.events, nil
}
func (eventQueryStore) Append(context.Context, eventstore.AppendRequest) ([]domain.Event, error) {
	return nil, nil
}
func (eventQueryStore) Scan(context.Context, string, int64, int) (eventstore.Page, error) {
	return eventstore.Page{}, nil
}
func (eventQueryStore) Event(context.Context, string) (domain.Event, error) {
	return domain.Event{}, nil
}

func TestBug003_QueriesKeepAllEvents(t *testing.T) {
	q := Queries{Store: eventQueryStore{events: []domain.Event{{Version: 1}, {Version: 2}}}}
	got, err := q.Events(context.Background(), "tenant", "order", "order-1")
	if err != nil || len(got) != 2 {
		t.Fatalf("events=%d err=%v", len(got), err)
	}
}
