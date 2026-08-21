package application

import (
	"context"
	"event-sourcing-service-test/internal/replay"
)

type Replay struct{ Jobs replay.Jobs }

func (r Replay) Start(c context.Context, t, p string) (string, error) { return r.Jobs.Create(c, t, p) }
