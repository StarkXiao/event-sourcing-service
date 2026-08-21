package application

import (
	"context"
	"event-sourcing-service-test/internal/consistency"
)

type Consistency struct{ Checker consistency.Checker }

func (s Consistency) Check(c context.Context, t, typ, id string) (string, error) {
	return s.Checker.Check(c, t, typ, id)
}
