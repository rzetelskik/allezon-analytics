package main

import (
	"context"
	"flag"
	"github.com/Shopify/sarama"
	"github.com/lovoo/goka"
	"github.com/rzetelskik/allezon-analytics/forwarder/internal/forwarder"
	"github.com/rzetelskik/allezon-analytics/shared/pkg/api"
	"github.com/rzetelskik/allezon-analytics/shared/pkg/kafka"
	"k8s.io/klog/v2"
	"log"
	"os"
	"runtime"
	"time"
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

	gokaConfig := goka.DefaultConfig()
	gokaConfig.Producer.Idempotent = true
	gokaConfig.Producer.RequiredAcks = sarama.WaitForAll
	gokaConfig.Net.MaxOpenRequests = 1

	tmc := goka.NewTopicManagerConfig()
	tmc.Table.Replication = 1
	tmc.Table.CleanupPolicy = "delete"
	tmc.Stream.Replication = 1
	tmc.Stream.CleanupPolicy = "delete"
	tmc.Stream.Retention = 24 * time.Hour
	tmc.MismatchBehavior = goka.TMConfigMismatchBehaviorFail

	tm, err := goka.NewTopicManager(bootstrap, gokaConfig, tmc)
	if err != nil {
		klog.Fatalf("can't create new goka topic manager: %v", err)
	}

	err = tm.EnsureStreamExists(string(kafka.UserProfileTopic), 1)
	if err != nil {
		klog.Fatalf("can't ensure stream \"%s\" exists: %v", kafka.UserProfileTopic, err)
	}

	err = tm.EnsureStreamExists(string(kafka.AggregateTopic), 1)
	if err != nil {
		klog.Fatalf("can't ensure stream \"%s\" exists: %v", kafka.AggregateTopic, err)
	}

	var group goka.Group = "forwarder"
	g := goka.DefineGroup(group,
		goka.Input(kafka.UserProfileTopic, new(api.UserTagCodec), forwarder.Forward),
		goka.Output(kafka.AggregateTopic, new(api.UserTagCodec)),
	)

	p, err := goka.NewProcessor(
		bootstrap,
		g,
		goka.WithTopicManagerBuilder(goka.TopicManagerBuilderWithTopicManagerConfig(tmc)),
		goka.WithConsumerGroupBuilder(goka.ConsumerGroupBuilderWithConfig(gokaConfig)),
	)
	if err != nil {
		klog.Fatalf("can't create new processor: %v", err)
	}

	err = p.Run(context.Background())
	if err != nil {
		klog.Fatalf("can't run processor: %v", err)
	}
}
