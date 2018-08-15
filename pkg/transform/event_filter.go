package transform

import (
	"sync"

	"k8s.io/apimachinery/pkg/watch"
)

type EventFilterPredicate func(event watch.Event) bool

type FilteredWatch struct {
	inputWatch watch.Interface
	resultChan chan watch.Event
	onceStop   sync.Once
	stop       chan struct{}
	predicates []EventFilterPredicate
}

func NewFilteredWatch(inputWatch watch.Interface, predicates ...EventFilterPredicate) *FilteredWatch {
	fw := &FilteredWatch{
		inputWatch: inputWatch,
		resultChan: make(chan watch.Event),
		predicates: predicates,
		stop:       make(chan struct{}),
	}
	go fw.readFilter()
	return fw
}

func (fw *FilteredWatch) readFilter() {
readLoop:
	for event := range fw.inputWatch.ResultChan() {
		for _, pred := range fw.predicates {
			if !pred(event) {
				continue readLoop
			}
		}
		fw.resultChan <- event
	}
}

func (fw *FilteredWatch) ResultChan() <-chan watch.Event {
	return fw.resultChan
}

func (fw *FilteredWatch) Stop() {
	fw.onceStop.Do(func() {
		fw.inputWatch.Stop()
		close(fw.resultChan)
	})
}
