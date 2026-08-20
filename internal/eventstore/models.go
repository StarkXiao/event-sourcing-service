package eventstore

import "event-sourcing-service/internal/domain"

type AppendRequest struct {
	TenantID, Type, ID, IdempotencyKey string
	ExpectedVersion                    int64
	Events                             []domain.Event
}
type Page struct {
	Events []domain.Event
	Next   int64
}
