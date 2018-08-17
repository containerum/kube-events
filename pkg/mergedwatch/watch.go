package mergedwatch

import (
	"sync"

	"k8s.io/apimachinery/pkg/watch"
)

type MergedWatch struct {
	watchers   []watch.Interface
	resultChan chan watch.Event

	onceStop sync.Once
}

func NewMergedWatch(watchers ...watch.Interface) *MergedWatch {
	mw := MergedWatch{
		watchers:   watchers,
		resultChan: make(chan watch.Event),
	}
	for _, w := range mw.watchers {
		go mw.retranslator(w)
	}
	return &mw
}

func (mw *MergedWatch) Stop() {
	mw.onceStop.Do(func() {
		for _, w := range mw.watchers {
			w.Stop()
		}
		close(mw.resultChan)
	})
}

func (mw *MergedWatch) retranslator(w watch.Interface) {
	for event := range w.ResultChan() {
		mw.resultChan <- event
	}
}

func (mw *MergedWatch) ResultChan() <-chan watch.Event {
	return mw.resultChan
}
