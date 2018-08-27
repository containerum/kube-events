package informerwatch

import (
	"sync"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

type InformerWatch struct {
	informer cache.SharedInformer

	onceStop   sync.Once
	stopChan   chan struct{}
	resultChan chan watch.Event
}

func NewInformerWatch(informer cache.SharedInformer) *InformerWatch {
	iw := &InformerWatch{
		informer:   informer,
		stopChan:   make(chan struct{}),
		resultChan: make(chan watch.Event),
	}
	iw.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			iw.resultChan <- watch.Event{
				Type:   watch.Added,
				Object: obj.(runtime.Object),
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			iw.resultChan <- watch.Event{
				Type:   watch.Modified,
				Object: newObj.(runtime.Object),
			}
		},
		DeleteFunc: func(obj interface{}) {
			iw.resultChan <- watch.Event{
				Type:   watch.Deleted,
				Object: obj.(runtime.Object),
			}
		},
	})
	go iw.informer.Run(iw.stopChan)
	return iw
}

func (iw *InformerWatch) ResultChan() <-chan watch.Event {
	return iw.resultChan
}

func (iw *InformerWatch) Stop() {
	iw.onceStop.Do(func() {
		close(iw.stopChan)
		close(iw.resultChan)
	})
}
