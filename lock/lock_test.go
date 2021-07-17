package lock

import (
	"context"
	"fmt"
	"sync"
	"testing"
)

func TestLockDepthRace(t *testing.T) {
	var data string
	wg := sync.WaitGroup{}
	fctx := Context(context.Background(), "firstctx")
	sctx := Context(context.Background(), "secondctx")
	l := New("lock")

	wg.Add(2)
	go func(ctx context.Context, data *string) {
		for i := 0; i < 1000; i++ {
			l.Lock(ctx)
			*data = fmt.Sprintf("%d", i)
		}
		for i := 0; i < 1000; i++ {
			l.Unlock(ctx)
		}
		wg.Done()
	}(fctx, &data)

	go func(ctx context.Context, data *string) {
		for i := 0; i < 1000; i++ {
			l.Lock(ctx)
			*data = fmt.Sprintf("%d", i)
		}
		for i := 0; i < 1000; i++ {
			l.Unlock(ctx)
		}
		wg.Done()
	}(sctx, &data)
	wg.Wait()
}
