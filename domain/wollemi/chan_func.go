package wollemi

import (
	"sync"
)

func NewChanFunc(workers, size int) *ChanFunc {
	this := &ChanFunc{}

	this.ch = make(chan func(), size)

	for i := 0; i < workers; i++ {
		this.wg.Add(1)
		go func() {
			defer this.wg.Done()
			for f := range this.ch {
				f()
			}
		}()
	}

	return this
}

type ChanFunc struct {
	wg sync.WaitGroup
	ch chan func()
}

func (this *ChanFunc) Close() {
	close(this.ch)
	this.wg.Wait()
}

func (this *ChanFunc) Run(f func()) {
	this.ch <- func() {
		f()
	}
}

func (this *ChanFunc) RunBlock(f func()) {
	var wg sync.WaitGroup

	wg.Add(1)

	this.Run(func() {
		f()
		wg.Done()
	})

	wg.Wait()
}
