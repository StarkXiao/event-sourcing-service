package application

import (
	"context"
	"testing"
	"time"
)

func TestBug005_WaitForProjectionHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	started := time.Now()
	err := (Queries{}).WaitForProjection(ctx)
	if err != context.Canceled || time.Since(started) > 50*time.Millisecond {
		t.Fatalf("err=%v", err)
	}
}
