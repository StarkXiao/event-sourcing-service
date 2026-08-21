package eventstore

import (
	"context"
	"event-sourcing-service-test/internal/domain"
)

type Store interface {
	Load(context.Context, string, string, string) ([]domain.Event, error)
	Append(context.Context, AppendRequest) ([]domain.Event, error)
	Scan(context.Context, string, int64, int) (Page, error)
	Event(context.Context, string) (domain.Event, error)
}
