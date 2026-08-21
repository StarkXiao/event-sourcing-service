package application

import (
	"errors"
	"event-sourcing-service-test/internal/domain"
	"testing"
)

func TestBug004_ProjectionLagIsIdentifiable(t *testing.T) {
	err := (Queries{}).ProjectionLag(2)
	if !errors.Is(err, domain.ErrProjectionLag) {
		t.Fatalf("error=%v", err)
	}
}
