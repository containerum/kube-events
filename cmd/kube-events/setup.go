package main

import (
	"time"

	kubeClientModel "github.com/containerum/kube-client/pkg/model"
	"github.com/containerum/kube-events/pkg/storage"
	"github.com/containerum/kube-events/pkg/storage/mongodb"
	"github.com/globalsign/mgo"
	log "github.com/sirupsen/logrus"
	"gopkg.in/urfave/cli.v2"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	configFlag = cli.StringFlag{
		Name:    "config",
		Aliases: []string{"c"},
		EnvVars: []string{"CONFIG"},
		Usage:   "Specify kubernetes config for connect. If not specified, use InClusterConfig for configuration.",
	}

	debugFlag = cli.BoolFlag{
		Name:    "debug",
		EnvVars: []string{"DEBUG"},
		Usage:   "Set log level to debug.",
	}

	textlogFlag = cli.BoolFlag{
		Name:    "textlog",
		EnvVars: []string{"TEXTLOG"},
		Usage:   "Print logs in text mode instead of json.",
	}

	mongoAddressFlag = cli.StringSliceFlag{
		Name:    "mongo_addr",
		EnvVars: []string{"MONGO_ADDR"},
		Usage:   "MongoDB host addresses.",
	}

	mongoUserFlag = cli.StringFlag{
		Name:    "MONGO_LOGIN",
		EnvVars: []string{"mongo_login"},
		Usage:   "Username to connect MongoDB.",
	}

	mongoPasswordFlag = cli.StringFlag{
		Name:    "mongo_password",
		EnvVars: []string{"MONGO_PASSWORD"},
		Usage:   "Password to connect MongoDB.",
	}

	mongoDatabaseFlag = cli.StringFlag{
		Name:    "mongo_db",
		EnvVars: []string{"MONGO_DB"},
		Usage:   "Database to use in MongoDB.",
		Value:   "kube-watches",
	}

	bufferCapacityFlag = cli.IntFlag{
		Name:    "buffer_capacity",
		EnvVars: []string{"BUFFER_CAPACITY"},
		Usage:   "Events buffer capacity (pre-allocated size).",
		Value:   200,
	}

	bufferMinInsertEventsFlag = cli.IntFlag{
		Name:    "buffer_min_insert_events",
		EnvVars: []string{"BUFFER_MIN_INSERT_EVENTS"},
		Usage:   "Minimal count of events in buffer to perform insert operation.",
		Value:   1,
	}

	bufferFlushPeriodFlag = cli.DurationFlag{
		Name:    "buffer_flush_period",
		EnvVars: []string{"BUFFER_FLUSH_PERIOD"},
		Usage:   "Events buffer to storage flush period.",
		Value:   30 * time.Second,
	}

	connectTimeoutFlag = cli.DurationFlag{
		Name:    "connection_timeout",
		EnvVars: []string{"CONNECTION_TIMEOUT"},
		Usage:   "Kubernetes connection timeout.",
		Value:   5 * time.Second,
	}
)

func setupLogs(ctx *cli.Context) {
	if ctx.Bool(debugFlag.Name) {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	if ctx.Bool(textlogFlag.Name) {
		log.SetFormatter(&log.TextFormatter{})
	} else {
		log.SetFormatter(&log.JSONFormatter{})
	}
}

func setupKubeClient(ctx *cli.Context) (*Kube, error) {
	var config *rest.Config
	var err error

	if cfg := ctx.String(configFlag.Name); cfg == "" {
		log.Info("Kube: Using InClusterConfig")
		config, err = rest.InClusterConfig()
	} else {
		log.Info("Kube: Using config from ", cfg)
		config, err = clientcmd.BuildConfigFromFlags("", cfg)
	}
	if err != nil {
		return nil, err
	}

	kubecli, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &Kube{
		Clientset: kubecli,
		config:    config,
	}, nil
}

type MongoLogrusAdapter struct {
	Log *log.Entry
}

func (ml *MongoLogrusAdapter) Output(callDepth int, s string) error {
	ml.Log.Debug(s)
	return nil
}

func setupMongo(ctx *cli.Context) (*mongodb.Storage, error) {
	mgo.SetDebug(ctx.Bool(debugFlag.Name))
	mgo.SetLogger(&MongoLogrusAdapter{Log: log.WithField("component", "mgo")})
	return mongodb.OpenConnection(&mgo.DialInfo{
		Addrs:     ctx.StringSlice(mongoAddressFlag.Name),
		Database:  ctx.String(mongoDatabaseFlag.Name),
		Mechanism: "SCRAM-SHA-1",
		Username:  ctx.String(mongoUserFlag.Name),
		Password:  ctx.String(mongoPasswordFlag.Name),
	})
}

func setupBuffer(ctx *cli.Context, inserter storage.EventBulkInserter, collector <-chan kubeClientModel.Event) *storage.RecordBuffer {
	return storage.NewRecordBuffer(storage.RecordBufferConfig{
		Storage:         inserter,
		BufferCap:       ctx.Int(bufferCapacityFlag.Name),
		InsertPeriod:    ctx.Duration(bufferFlushPeriodFlag.Name),
		MinInsertEvents: ctx.Int(bufferMinInsertEventsFlag.Name),
		Collector:       collector,
	})
}
