package lock

import (
	"context"
	"sync"
)

type LockFn func(ctx context.Context)
type LockKey string

type Lock struct {
	id     string
	lockid string
	depth  uint32
	cnd    *sync.Cond
}

func Context(parent context.Context, id string) context.Context {
	return context.WithValue(parent, LockKey("id"), id)
}

func New(id string) *Lock {
	return &Lock{
		id:  id,
		cnd: sync.NewCond(&sync.Mutex{}),
	}
}

func (l *Lock) Lock(ctx context.Context) bool {
	l.cnd.L.Lock()
	for {
		if l.doLock(ctx) {
			l.cnd.L.Unlock()
			l.cnd.Signal()
			return true
		}
		l.cnd.Wait()
	}
}

func (l *Lock) Unlock(ctx context.Context) bool {
	l.cnd.L.Lock()
	defer l.cnd.L.Unlock()
	defer l.cnd.Signal()
	return l.doUnlock(ctx)
}

func (l *Lock) doLock(ctx context.Context) bool {
	switch {
	case l.depth == 0:
		l.lockid = ctx.Value(LockKey("id")).(string)
		l.depth++
		return true
	case l.lockid != ctx.Value(LockKey("id")):
		return false
	default:
		l.depth++
		return true
	}
}

func (l *Lock) doUnlock(ctx context.Context) bool {
	switch {
	case l.lockid == ctx.Value(LockKey("id")):
		l.depth--
		if l.depth == 0 {
			l.lockid = ""
			return true
		}
		return false
	case l.depth == 0:
		l.lockid = ""
		return true
	default:
		return false
	}
}
