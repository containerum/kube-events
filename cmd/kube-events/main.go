package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"

	"github.com/containerum/kube-events/pkg/model"
	"github.com/containerum/kube-events/pkg/transform"
	"gopkg.in/urfave/cli.v2"
	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

var defaultListOptions = meta_v1.ListOptions{
	Watch: true,
}

var eventTransformer = transform.EventTransformer{
	RuleSelector: func(event watch.Event) string {
		return string(ObservableTypeFromObject(event.Object))
	},
	Rules: map[string]transform.Func{
		string(model.ObservableNamespace):        MakeNamespaceRecord,
		string(model.ObservableDeployment):       MakeDeployRecord,
		string(model.ObservablePod):              MakePodRecord,
		string(model.ObservableService):          MakeServiceRecord,
		string(model.ObservableIngress):          MakeIngressRecord,
		string(model.ObservablePersistentVolume): MakePVRecord,
		string(model.ObservableNode):             MakeNodeRecord,
	},
}

func action(ctx *cli.Context) error {
	client, err := setupKubeClient(ctx)
	if err != nil {
		return err
	}

	watcher, err := client.CoreV1().Events("").Watch(defaultListOptions)
	if err != nil {
		return err
	}

	go func() {
		sigch := make(chan os.Signal)
		signal.Notify(sigch, os.Interrupt, os.Kill)
		<-sigch
		watcher.Stop()
	}()
	defer watcher.Stop()

	for event := range watcher.ResultChan() {
		switch event.Type {
		case watch.Added, watch.Deleted, watch.Modified:
			str, err := json.Marshal(event.Object.(*core_v1.Event))
			if err != nil {
				return err
			}
			fmt.Printf("%s: %s\n", event.Type, str)
		default:
			fmt.Printf("%#v\n", event)
		}
	}

	return nil
}

func main() {
	app := cli.App{
		Name:        "kube-events",
		Description: "Subscribes for kubernetes watches, filters it and records to storage.",
		Flags: []cli.Flag{
			&configFlag,
			&debugFlag,
			&textlogFlag,
			&retentionPeriodFlag,
			&cleanupIntervalFlag,
			&mongoAddressFlag,
			&mongoUserFlag,
			&mongoPasswordFlag,
			&mongoDatabaseFlag,
			&bufferCapacityFlag,
			&bufferFlushPeriodFlag,
			&bufferMinInsertEventsFlag,
		},
		Action: action,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}
}
