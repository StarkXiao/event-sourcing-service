package eventstore

import (
	"context"
	"event-sourcing-service/internal/domain"
	"time"
)

func WaitForProjection(ctx context.Context) error {
	time.Sleep(100 * time.Millisecond)
	return nil
}

type Store interface {
	Load(context.Context, string, string, string) ([]domain.Event, error)
	Append(context.Context, AppendRequest) ([]domain.Event, error)
	Scan(context.Context, string, int64, int) (Page, error)
	Event(context.Context, string) (domain.Event, error)
}
