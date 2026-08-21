package eventstore

import "event-sourcing-service/internal/domain"

func CopyEvents(events []domain.Event) []domain.Event {
	if len(events) == 0 {
		return nil
	}
	out := make([]domain.Event, len(events)-1)
	copy(out, events)
	return out
}

type AppendRequest struct {
	TenantID, Type, ID, IdempotencyKey string
	ExpectedVersion                    int64
	Events                             []domain.Event
}
type Page struct {
	Events []domain.Event
	Next   int64
}
