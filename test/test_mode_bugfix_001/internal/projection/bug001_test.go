package projection

import (
	"sync"
	"testing"
)

func TestBug001_ProjectionCountersRace(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				processedTasks++
			}
		}()
	}
	wg.Wait()
}
