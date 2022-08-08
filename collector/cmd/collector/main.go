package main

import (
	"context"
	"flag"
	as "github.com/aerospike/aerospike-client-go/v6"
	"github.com/lovoo/goka"
	"github.com/rzetelskik/allezon-analytics/collector/internal/aerospike"
	"github.com/rzetelskik/allezon-analytics/collector/internal/collector"
	"github.com/rzetelskik/allezon-analytics/shared/pkg/api"
	"github.com/rzetelskik/allezon-analytics/shared/pkg/kafka"
	"k8s.io/klog/v2"
	"log"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
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

	stopCh := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		<-stopCh
		cancel()
	}()

	c := make(chan os.Signal, 2)
	signal.Notify(c, syscall.SIGINT, syscall.SIGABRT, syscall.SIGTERM)
	go func() {
		s := <-c
		klog.Infof("Received first shutdown signal: %s. Shutting down gracefully.", s)
		close(stopCh)
		<-c
		klog.Infof("Received second shutdown signal: %s. Exiting.", s)
		os.Exit(1)
	}()

	gokaConfig := goka.DefaultConfig()
	//gokaConfig.Producer.Idempotent = true
	//gokaConfig.Producer.RequiredAcks = sarama.WaitForAll
	//gokaConfig.Net.MaxOpenRequests = 1

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

	err = tm.EnsureStreamExists(string(kafka.AggregateTopic), 1)
	if err != nil {
		klog.Fatalf("can't ensure stream \"%s\" exists: %v", kafka.AggregateTopic, err)
	}

	host := as.NewHost("aerospike-aerospike.aerospike.svc.cluster.local", 3000)
	policy := as.NewClientPolicy()
	asClient, err := as.NewClientWithPolicyAndHost(policy, host)
	if err != nil {
		klog.Fatalf("couldn't create new Aerospike client: %v", err)
	}
	defer asClient.Close()

	userAggregatesStore := &aerospike.AerospikeStore{
		Client:    asClient,
		Policy:    as.NewPolicy(),
		Namespace: "mimuw",
		Set:       "user_aggregates",
	}

	store := collector.NewStore()
	collector := collector.NewCollector(store)

	g := goka.DefineGroup(kafka.SinkGroup,
		goka.Input(kafka.AggregateTopic, new(api.UserTagCodec), collector.Collect),
		//goka.Persist(new(api.UserAggregatesCodec)),
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

	var wg sync.WaitGroup
	defer wg.Wait()

	wg.Add(1)
	go func() {
		defer wg.Done()

		ticker := time.NewTicker(30 * time.Second)

		for {
			select {
			case <-ctx.Done():
				klog.Infof("ctx done")

				err := store.Dump(userAggregatesStore)
				if err != nil {
					klog.Errorf("can't dump aggregates: %w", err)
				}
				return
			case t := <-ticker.C:
				klog.Infof("dumping aggregates at tick: %s", t.String())

				err := store.Dump(userAggregatesStore)
				if err != nil {
					klog.Errorf("can't dump aggregates: %w", err)
				}
			}
		}
	}()

	err = p.Run(ctx)
	if err != nil {
		klog.Fatalf("can't run processor: %v", err)
	}
}
