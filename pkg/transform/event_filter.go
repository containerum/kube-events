package transform

import "k8s.io/apimachinery/pkg/watch"

type EventFilterPredicate func(event watch.Event) bool

type EventFilter struct {
	Predicates []EventFilterPredicate
}

func (ef *EventFilter) Output(input <-chan watch.Event) <-chan watch.Event {
	outCh := make(chan watch.Event)
	go func() {
	eventRead:
		for event := range input {
			for _, pred := range ef.Predicates {
				if !pred(event) {
					continue eventRead
				}
			}
			outCh <- event
		}
		close(outCh)
	}()
	return outCh
}
