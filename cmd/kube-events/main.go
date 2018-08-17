package main

import (
	"fmt"
	"os"
	"os/signal"
	"text/tabwriter"

	"github.com/containerum/kube-events/pkg/model"
	"github.com/containerum/kube-events/pkg/transform"
	"gopkg.in/urfave/cli.v2"
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

func printFlags(ctx *cli.Context) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.TabIndent|tabwriter.Debug)
	for _, f := range ctx.FlagNames() {
		fmt.Fprintf(w, "Flag: %s\t Value: %s\n", f, ctx.String(f))
	}
	return w.Flush()
}

func action(ctx *cli.Context) error {
	setupLogs(ctx)

	kubeClient, err := setupKubeClient(ctx)
	if err != nil {
		return err
	}
	watcher, err := kubeClient.WatchSupportedResources(defaultListOptions)
	if err != nil {
		return err
	}
	defer watcher.Stop()

	mongoStorage, err := setupMongo(ctx)
	if err != nil {
		return err
	}
	defer mongoStorage.Close()

	eventBuffer, err := setupBuffer(ctx, mongoStorage, eventTransformer.Output(watcher.ResultChan()))
	if err != nil {
		return err
	}
	defer eventBuffer.Stop()
	go eventBuffer.RunCollection()

	sigch := make(chan os.Signal)
	signal.Notify(sigch, os.Kill, os.Interrupt)
	<-sigch

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
			&mongoCollectionSizeFlag,
			&mongoCollectionMaxDocsFlag,
			&bufferCapacityFlag,
			&bufferFlushPeriodFlag,
			&bufferMinInsertEventsFlag,
		},
		Before: printFlags,
		Action: action,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}
}
