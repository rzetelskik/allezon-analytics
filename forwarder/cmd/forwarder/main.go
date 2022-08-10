package main

import (
	"context"
	"flag"
	"github.com/lovoo/goka"
	"github.com/rzetelskik/allezon-analytics/forwarder/internal/forwarder"
	"github.com/rzetelskik/allezon-analytics/shared/pkg/api"
	"github.com/rzetelskik/allezon-analytics/shared/pkg/kafka"
	"k8s.io/klog/v2"
	"log"
	"os"
	"runtime"
)

var (
	bootstrap = []string{"kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"}
)

func main() {
	var err error

	klog.InitFlags(flag.CommandLine)
	err = flag.Set("logtostderr", "true")
	if err != nil {
		panic(err)
	}
	flag.Parse()
	defer klog.Flush()

	runtime.GOMAXPROCS(runtime.NumCPU())
	log.SetOutput(os.Stdout)

	var group goka.Group = "forwarder"
	g := goka.DefineGroup(group,
		goka.Input(kafka.UserProfileTopic, new(api.UserTagCodec), forwarder.Forward),
		goka.Output(kafka.AggregateTopic, new(api.UserTagCodec)),
	)

	p, err := goka.NewProcessor(
		bootstrap,
		g,
	)
	if err != nil {
		klog.Fatalf("can't create new processor: %v", err)
	}

	err = p.Run(context.Background())
	if err != nil {
		klog.Fatalf("can't run processor: %v", err)
	}
}
