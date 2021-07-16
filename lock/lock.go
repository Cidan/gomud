package lock

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type LockFn func(ctx context.Context)

type Lock struct {
	id     string
	locker uint32
	chk    sync.Mutex
	// This context is used for reentrant locking.
	ctx *context.Context
}

func New(id string) *Lock {
	return &Lock{
		id:  id,
		chk: sync.Mutex{},
	}
}

func (l *Lock) Lock(ctx context.Context) bool {
	for {
		if l.doLock(ctx) {
			return true
		}
		time.Sleep(1 * time.Millisecond)
	}
}

func (l *Lock) Unlock(ctx context.Context) bool {
	return l.doUnlock(ctx)
}

func (l *Lock) TryLock(ctx context.Context, fn LockFn) {
	if l.doLock(ctx) {
		defer l.doUnlock(ctx)
		fn(ctx)
	}
}

func (l *Lock) doLock(ctx context.Context) bool {
	l.chk.Lock()
	defer l.chk.Unlock()
	if atomic.CompareAndSwapUint32(&l.locker, 0, 1) || l.ctx == &ctx {
		l.ctx = &ctx
		return true
	}
	return false
}

func (l *Lock) doUnlock(ctx context.Context) bool {
	l.chk.Lock()
	defer l.chk.Unlock()
	if !atomic.CompareAndSwapUint32(&l.locker, 1, 0) || l.ctx != &ctx {
		return false
	}
	l.ctx = nil
	return true
}
