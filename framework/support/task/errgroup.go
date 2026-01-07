package task

import (
	"context"
	"github.com/feimumoke/wechating/framework/wcerror"
	"sync"
)

type errGroup struct {
	cancel  func()
	wg      sync.WaitGroup
	errOnce sync.Once
	err     *bmserror.BMSError
}

func NewErrGroupWithContext(ctx context.Context) (*errGroup, context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	return &errGroup{cancel: cancel}, ctx
}

func (g *errGroup) Wait() *bmserror.BMSError {
	g.wg.Wait()
	if g.cancel != nil {
		g.cancel()
	}
	return g.err
}

// 在go routine内调用，调用之前必须在协程外调用Add方法
func (g *errGroup) Go(f func() *bmserror.BMSError) {
	// g.wg.Add(1)
	defer g.wg.Done()
	if err := f(); err != nil {
		g.errOnce.Do(func() {
			g.err = err
			if g.cancel != nil {
				g.cancel()
			}
		})
	}
}

func (g *errGroup) Add() {
	g.wg.Add(1)
}
