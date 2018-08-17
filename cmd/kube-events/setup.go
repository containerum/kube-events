package main

import (
	"fmt"
	"time"

	"github.com/containerum/kube-events/pkg/model"
	"github.com/containerum/kube-events/pkg/storage"
	"github.com/containerum/kube-events/pkg/storage/mongodb"
	"github.com/globalsign/mgo"
	"github.com/sirupsen/logrus"
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
		EnvVars: []string{"TEXT_LOG"},
		Usage:   "Print logs in text mode instead of json.",
	}

	retentionPeriodFlag = cli.DurationFlag{
		Name:    "retention-period",
		EnvVars: []string{"RETENTION_PERIOD"},
		Usage:   "Period of keeping log records in storage. Earlier records will be deleted.",
		Value:   24 * time.Hour,
	}

	cleanupIntervalFlag = cli.DurationFlag{
		Name:    "cleanup-period",
		EnvVars: []string{"CLEANUP_PERIOD"},
		Usage:   "Period of running cleanup procedure.",
		Value:   12 * time.Hour,
	}

	mongoAddressFlag = cli.StringSliceFlag{
		Name:    "mongo-address",
		EnvVars: []string{"MONGO_ADDRS"},
		Usage:   "MongoDB host addresses.",
	}

	mongoUserFlag = cli.StringFlag{
		Name:    "mongo-user",
		EnvVars: []string{"MONGO_USER"},
		Usage:   "Username to connect MongoDB.",
	}

	mongoPasswordFlag = cli.StringFlag{
		Name:    "mongo-password",
		EnvVars: []string{"MONGO_PASSWORD"},
		Usage:   "Password to connect MongoDB.",
	}

	mongoDatabaseFlag = cli.StringFlag{
		Name:    "mongo-database",
		EnvVars: []string{"MONGO_DATABASE"},
		Usage:   "Database to use in MongoDB.",
		Value:   "kube-watches",
	}

	bufferCapacityFlag = cli.IntFlag{
		Name:    "buffer-capacity",
		EnvVars: []string{"BUFFER_CAPACITY"},
		Usage:   "Events buffer capacity (pre-allocated size).",
		Value:   500,
	}

	bufferMinInsertEventsFlag = cli.IntFlag{
		Name:    "buffer-min-insert-events",
		EnvVars: []string{"BUFFER_MIN_INSERT_EVENTS"},
		Usage:   "Minimal count of events in buffer to perform insert operation.",
		Value:   10,
	}

	bufferFlushPeriodFlag = cli.DurationFlag{
		Name:    "buffer-flush-period",
		EnvVars: []string{"BUFFER_FLUSH_PERIOD"},
		Usage:   "Events buffer to storage flush period.",
		Value:   30 * time.Second,
	}
)

func setupLogs(ctx *cli.Context) {
	if ctx.Bool(debugFlag.Name) {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	if ctx.Bool(textlogFlag.Name) {
		logrus.SetFormatter(&logrus.TextFormatter{})
	} else {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}
}

func setupKubeClient(ctx *cli.Context) (*Kube, error) {
	var config *rest.Config
	var err error

	if cfg := ctx.String(configFlag.Name); cfg == "" {
		fmt.Println("Using InClusterConfig")
		config, err = rest.InClusterConfig()
	} else {
		fmt.Println("Using config from", cfg)
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

func setupMongo(ctx *cli.Context) (*mongodb.Storage, error) {
	return mongodb.OpenConnection(&mongodb.Config{
		DialInfo: mgo.DialInfo{
			Addrs:     ctx.StringSlice(mongoAddressFlag.Name),
			Database:  ctx.String(mongoDatabaseFlag.Name),
			Mechanism: "SCRAM-SHA-1",
			Username:  ctx.String(mongoUserFlag.Name),
			Password:  ctx.String(mongoPasswordFlag.Name),
		},
	})
}

func setupBuffer(ctx *cli.Context, inserter storage.EventBulkInserter, collector <-chan model.Record) (*storage.RecordBuffer, error) {
	return storage.NewRecordBuffer(storage.RecordBufferConfig{
		Storage:         inserter,
		BufferCap:       ctx.Int(bufferCapacityFlag.Name),
		InsertPeriod:    ctx.Duration(bufferFlushPeriodFlag.Name),
		MinInsertEvents: ctx.Int(bufferMinInsertEventsFlag.Name),
		Collector:       collector,
	}), nil
}

func setupCleaner(ctx *cli.Context, cleaner storage.EventCleaner) (*storage.RecordCleaner, error) {
	return storage.NewRecordCleaner(storage.RecordCleanerConfig{
		Storage:          cleaner,
		CleanupRunPeriod: ctx.Duration(cleanupIntervalFlag.Name),
		RetentionPeriod:  ctx.Duration(retentionPeriodFlag.Name),
	}), nil
}
