package transform

import (
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/watch"

	kubeClientModel "github.com/containerum/kube-client/pkg/model"
)

type RuleSelector func(event watch.Event) string

type Func func(event watch.Event) kubeClientModel.Event

type EventTransformer struct {
	RuleSelector RuleSelector
	Rules        map[string]Func
}

func (et *EventTransformer) Output(input <-chan watch.Event) <-chan kubeClientModel.Event {
	outCh := make(chan kubeClientModel.Event)
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
