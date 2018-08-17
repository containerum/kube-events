package transform

import (
	"github.com/containerum/kube-events/pkg/model"
	"k8s.io/apimachinery/pkg/watch"
)

type RuleSelector func(event watch.Event) string

type Func func(event watch.Event) model.Record

type EventTransformer struct {
	RuleSelector RuleSelector
	Rules        map[string]Func
}

func (et *EventTransformer) Output(input <-chan watch.Event) <-chan model.Record {
	outCh := make(chan model.Record)
	go func() {
		for event := range input {
			f, ok := et.Rules[et.RuleSelector(event)]
			if !ok {
				continue
			}
			outCh <- f(event)
		}
		close(outCh)
	}()
	return outCh
}
