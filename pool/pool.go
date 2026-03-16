package pool

import (
	"sync"
	"sync/atomic"

	"emperror.dev/errors"
	"github.com/sourcegraph/conc/panics"
)

type GoroutinePool struct {
	maxCapacity atomic.Int32
}

func (p *GoroutinePool) Submit(task func()) {
	if p.maxCapacity.Load() <= 1 {
		task()
	} else {
		p.maxCapacity.Add(-1)
		go func() {
			defer p.maxCapacity.Add(1)
			task()
		}()
	}
}

func NewGoroutinePool(maxCapacity int) *GoroutinePool {
	p := &GoroutinePool{}
	p.maxCapacity.Store(int32(maxCapacity))
	return p
}

type GOPool struct {
	pool *GoroutinePool
}

func NewGOPool(size int) *GOPool {
	newPool := NewGoroutinePool(size)
	return &GOPool{pool: newPool}
}

func (p *GOPool) WaitGOIndex(size int, f func(index int) error) (err error) {
	waitGroup := new(sync.WaitGroup)
	waitGroup.Add(size)
	for i := 0; i < size; i++ {
		index := i
		p.pool.Submit(func() {
			defer waitGroup.Done()
			recovered := panics.Try(func() {
				err1 := f(index)
				if err1 != nil && err == nil {
					err = errors.WithStackIf(err1)
				}
			})
			err1 := recovered.AsError()
			if err1 != nil && err == nil {
				err = errors.WithStackIf(err1)
			}
		})
	}
	waitGroup.Wait()
	return
}

func (p *GOPool) WaitGO(f func() error) (err error) {
	var wg sync.WaitGroup
	wg.Add(1)
	p.pool.Submit(func() {
		recovered := panics.Try(func() {
			defer wg.Done()
			err2 := f()
			if err2 != nil && err == nil {
				err = errors.WithStackIf(err2)
			}
		})
		err1 := recovered.AsError()
		if err1 != nil && err == nil {
			err = errors.WithStackIf(err1)
		}
	})
	wg.Wait()
	return
}
