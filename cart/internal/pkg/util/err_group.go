package util

import (
	"context"
	"sync"
)

type ErrGroup struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	errOnce sync.Once
	err     error
}

func ErrGroupWithContext(ctx context.Context) (*ErrGroup, context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	return &ErrGroup{ctx: ctx, cancel: cancel}, ctx
}

func (eg *ErrGroup) Go(task func() error) {
	select {
	case <-eg.ctx.Done():
		return
	default:
	}

	eg.wg.Add(1)

	go func() {
		defer eg.wg.Done()

		if err := task(); err != nil {
			eg.errOnce.Do(func() {
				eg.err = err
				eg.cancel()
			})
		}
	}()
}

func (eg *ErrGroup) Wait() error {
	eg.wg.Wait()
	if eg.cancel != nil {
		eg.cancel()
	}

	return eg.err
}
