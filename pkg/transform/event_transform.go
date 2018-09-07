package transform

import (
	"github.com/containerum/kube-events/pkg/model"
	"github.com/sirupsen/logrus"
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
			ruleSelector := et.RuleSelector(event)
			f, ok := et.Rules[ruleSelector]
			if !ok {
				logrus.Warn("Unsupported RuleSelector: %v", ruleSelector)
				continue
			}
			outCh <- f(event)
		}
		close(outCh)
	}()
	return outCh
}
