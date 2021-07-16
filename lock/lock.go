package lock

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
)

type LockFn func(ctx context.Context)
type LockKey string

type Lock struct {
	id     string
	lockid string
	locker uint32
	chk    sync.Mutex
}

func Context(parent context.Context, id string) context.Context {
	return context.WithValue(parent, LockKey("id"), id)
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
		// Allow other threads to run so we don't deadlock.
		runtime.Gosched()
		//time.Sleep(1 * time.Millisecond)
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

	if atomic.CompareAndSwapUint32(&l.locker, 0, 1) && (l.lockid == "" || l.lockid == ctx.Value(LockKey("id"))) {
		l.lockid = ctx.Value(LockKey("id")).(string)
		return true
	}
	atomic.CompareAndSwapUint32(&l.locker, 1, 0)
	return false
}

func (l *Lock) doUnlock(ctx context.Context) bool {
	l.chk.Lock()
	defer l.chk.Unlock()
	if !atomic.CompareAndSwapUint32(&l.locker, 1, 0) || l.lockid != ctx.Value(LockKey("id")) {
		return false
	}
	l.lockid = ""
	return true
}
