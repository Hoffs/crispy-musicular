package syncplus

import (
	"context"
	"sync"
)

// returns true if timed out
func WaitContext(ctx context.Context, wg *sync.WaitGroup) bool {
	c := make(chan struct{})
	go func() {
		// This potentially leaks.
		defer close(c)
		wg.Wait()
	}()

	select {
	case <-c:
		return false
	case <-ctx.Done():
		return true
	}
}
