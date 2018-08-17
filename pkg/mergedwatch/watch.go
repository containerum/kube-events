package mergedwatch

import (
	"sync"

	"k8s.io/apimachinery/pkg/watch"
)

type MergedWatch struct {
	watchers   []watch.Interface
	resultChan chan watch.Event

	onceStop  sync.Once
	onceClose sync.Once
}

func NewMergedWatch(watchers ...watch.Interface) *MergedWatch {
	mw := MergedWatch{
		watchers:   watchers,
		resultChan: make(chan watch.Event),
	}
	var wg sync.WaitGroup
	wg.Add(len(watchers))
	for _, w := range mw.watchers {
		go mw.retranslator(w, &wg)
		go mw.waitClose(&wg)
	}
	return &mw
}

func (mw *MergedWatch) Stop() {
	mw.onceStop.Do(func() {
		for _, w := range mw.watchers {
			w.Stop()
		}
	})
}

func (mw *MergedWatch) retranslator(w watch.Interface, wg *sync.WaitGroup) {
	for event := range w.ResultChan() {
		mw.resultChan <- event
	}
	wg.Done()
}

func (mw *MergedWatch) waitClose(wg *sync.WaitGroup) {
	wg.Wait()
	mw.onceClose.Do(func() {
		close(mw.resultChan)
	})
}

func (mw *MergedWatch) ResultChan() <-chan watch.Event {
	return mw.resultChan
}
