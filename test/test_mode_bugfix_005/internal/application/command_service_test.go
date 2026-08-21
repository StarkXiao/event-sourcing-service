package application

import (
	"context"
	"event-sourcing-service-test/internal/domain"
	"event-sourcing-service-test/internal/eventstore"
	"testing"
)

type commandStore struct {
	events []domain.Event
}

func (s *commandStore) Load(context.Context, string, string, string) ([]domain.Event, error) {
	return s.events, nil
}

func (s *commandStore) Append(_ context.Context, r eventstore.AppendRequest) ([]domain.Event, error) {
	s.events = append(s.events, r.Events...)
	return r.Events, nil
}

func (*commandStore) Scan(context.Context, string, int64, int) (eventstore.Page, error) {
	return eventstore.Page{}, nil
}

func (*commandStore) Event(context.Context, string) (domain.Event, error) {
	return domain.Event{}, nil
}

func TestExecuteRejectsMissingAccountCurrency(t *testing.T) {
	s := &commandStore{}
	_, err := (Commands{Store: s}).Execute(context.Background(), "tenant", "account", "account-1", "open", "key", 0, map[string]any{"currency": nil})
	if err != domain.ErrInvalid {
		t.Fatalf("error = %v, want %v", err, domain.ErrInvalid)
	}
}

func TestExecuteRejectsFractionalAmount(t *testing.T) {
	s := &commandStore{events: []domain.Event{{Type: domain.AccountOpened, Payload: map[string]any{"currency": "USD"}}}}
	_, err := (Commands{Store: s}).Execute(context.Background(), "tenant", "account", "account-1", "credit", "key", 1, map[string]any{"amount": 1.5, "currency": "USD"})
	if err != domain.ErrInvalid {
		t.Fatalf("error = %v, want %v", err, domain.ErrInvalid)
	}
}
