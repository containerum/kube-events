package mergedwatch

import (
	"sync"

	"k8s.io/apimachinery/pkg/watch"
)

type MergedWatch struct {
	watchers   []watch.Interface
	resultChan chan watch.Event

	closedMutex sync.Mutex
	closed      bool
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
	mw.closedMutex.Lock()
	defer mw.closedMutex.Unlock()
	if !mw.closed {
		for _, w := range mw.watchers {
			w.Stop()
		}
		close(mw.resultChan)
		mw.closed = true
	}
}

func (mw *MergedWatch) retranslator(w watch.Interface) {
	for event := range w.ResultChan() {
		mw.resultChan <- event
	}
}

func (mw *MergedWatch) ResultChan() <-chan watch.Event {
	return mw.resultChan
}
