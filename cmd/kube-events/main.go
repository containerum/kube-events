package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"text/tabwriter"
	"time"

	"github.com/containerum/kube-events/pkg/storage/mongodb"

	"github.com/containerum/kube-events/pkg/model"
	"github.com/containerum/kube-events/pkg/transform"
	log "github.com/sirupsen/logrus"
	"gopkg.in/urfave/cli.v2"
	"k8s.io/apimachinery/pkg/watch"
)

var eventTransformer = transform.EventTransformer{
	RuleSelector: func(event watch.Event) string {
		return string(ObservableTypeFromObject(event.Object))
	},
	Rules: map[string]transform.Func{
		string(model.ObservableNamespace):             MakeNamespaceRecord,
		string(model.ObservableDeployment):            MakeDeployRecord,
		string(model.ObservableEvent):                 MakeEventRecord,
		string(model.ObservableService):               MakeServiceRecord,
		string(model.ObservableIngress):               MakeIngressRecord,
		string(model.ObservablePersistentVolumeClaim): MakePVCRecord,
		string(model.ObservableNode):                  MakeNodeRecord,
		string(model.ObservableSecret):                MakeSecretRecord,
		string(model.ObservableConfigMap):             MakeConfigMapRecord,
	},
}

func printFlags(ctx *cli.Context) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.TabIndent|tabwriter.Debug)
	for _, f := range ctx.FlagNames() {
		fmt.Fprintf(w, "Flag: %s\t Value: %s\n", f, ctx.String(f))
	}
	return w.Flush()
}

func pingKube(client *Kube, pingPeriod time.Duration, errChan chan<- error, stopChan <-chan struct{}) {
	ticker := time.NewTicker(pingPeriod)
	httpClient := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: client.config.Timeout,
	}
	reqURL, err := url.Parse(client.config.Host)
	if err != nil {
		errChan <- err
		return
	}
	reqURL.Path = "/healthz"
	req := http.Request{
		Method: http.MethodGet,
		URL:    reqURL,
	}
	defer ticker.Stop()
	for {
		select {
		case <-stopChan:
			return
		case <-ticker.C:
			log.Debug("Ping kube ", req.URL)
			resp, err := httpClient.Do(&req)
			if err != nil {
				errChan <- err
				continue
			}
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				errChan <- err
				continue
			}
			if err := resp.Body.Close(); err != nil {
				errChan <- err
				continue
			}
			if resp.StatusCode != http.StatusOK || string(body) != "ok" {
				errChan <- fmt.Errorf("%s", body)
			}
		}
	}
}

func action(ctx *cli.Context) error {
	setupLogs(ctx)

	kubeClient, err := setupKubeClient(ctx)
	if err != nil {
		return err
	}

	watchers := kubeClient.WatchSupportedResources()

	mongoStorage, err := setupMongo(ctx)
	if err != nil {
		return err
	}
	defer mongoStorage.Close()

	//Namespaces
	defer watchers.ResourceQuotas.Stop()
	nsBuffer := setupBuffer(ctx, mongoStorage, eventTransformer.Output(watchers.ResourceQuotas.ResultChan()))
	defer nsBuffer.Stop()
	go nsBuffer.RunCollection(mongodb.ResourceQuotasCollection)

	//Deployments
	defer watchers.Deployments.Stop()
	deplBuffer := setupBuffer(ctx, mongoStorage, eventTransformer.Output(watchers.Deployments.ResultChan()))
	defer deplBuffer.Stop()
	go deplBuffer.RunCollection(mongodb.DeploymentCollection)

	//Services
	defer watchers.Services.Stop()
	svcBuffer := setupBuffer(ctx, mongoStorage, eventTransformer.Output(watchers.Services.ResultChan()))
	defer svcBuffer.Stop()
	go svcBuffer.RunCollection(mongodb.ServiceCollection)

	//Ingresses
	defer watchers.Ingresses.Stop()
	ingrBuffer := setupBuffer(ctx, mongoStorage, eventTransformer.Output(watchers.Ingresses.ResultChan()))
	defer ingrBuffer.Stop()
	go ingrBuffer.RunCollection(mongodb.IngressCollection)

	//Volumes
	defer watchers.PVCs.Stop()
	pvcBuffer := setupBuffer(ctx, mongoStorage, eventTransformer.Output(watchers.PVCs.ResultChan()))
	defer pvcBuffer.Stop()
	go pvcBuffer.RunCollection(mongodb.PVCCollection)

	//Secrets
	defer watchers.PVCs.Stop()
	secretBuffer := setupBuffer(ctx, mongoStorage, eventTransformer.Output(watchers.Secrets.ResultChan()))
	defer secretBuffer.Stop()
	go secretBuffer.RunCollection(mongodb.SecretsCollection)

	//ConfigMaps
	defer watchers.PVCs.Stop()
	cmBuffer := setupBuffer(ctx, mongoStorage, eventTransformer.Output(watchers.ConfigMaps.ResultChan()))
	defer cmBuffer.Stop()
	go cmBuffer.RunCollection(mongodb.ConfigMapsCollection)

	//Events
	defer watchers.Events.Stop()
	eventBuffer := setupBuffer(ctx, mongoStorage, eventTransformer.Output(watchers.Events.ResultChan()))
	defer eventBuffer.Stop()
	go eventBuffer.RunCollection(mongodb.EventsCollection)

	pingStopChan := make(chan struct{})
	defer close(pingStopChan)
	pingErrChan := make(chan error)
	go pingKube(kubeClient, 5*time.Second, pingErrChan, pingStopChan)

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)
	select {
	case <-sigch:
		return nil
	case err := <-pingErrChan:
		log.WithError(err).Errorf("Ping kube failed")
		os.Exit(1)
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
			&mongoAddressFlag,
			&mongoUserFlag,
			&mongoPasswordFlag,
			&mongoDatabaseFlag,
			&bufferCapacityFlag,
			&bufferFlushPeriodFlag,
			&bufferMinInsertEventsFlag,
			&connectTimeoutFlag,
		},
		Before: printFlags,
		Action: action,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}
}
