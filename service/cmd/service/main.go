package main

import (
	"context"
	"errors"
	"flag"
	as "github.com/aerospike/aerospike-client-go/v6"
	"github.com/lovoo/goka"
	"github.com/lovoo/goka/codec"
	"github.com/rzetelskik/allezon-analytics/service/internal/service/aerospike"
	"github.com/rzetelskik/allezon-analytics/service/internal/service/server"
	"github.com/rzetelskik/allezon-analytics/shared/pkg/api"
	"github.com/rzetelskik/allezon-analytics/shared/pkg/kafka"
	"k8s.io/klog/v2"
	"log"
	"net/http"
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

	host := as.NewHost("aerospike-aerospike.aerospike.svc.cluster.local", 3000)
	policy := as.NewClientPolicy()
	asClient, err := as.NewClientWithPolicyAndHost(policy, host)
	if err != nil {
		klog.Fatalf("couldn't create new Aerospike client: %v", err)
	}
	defer asClient.Close()

	userProfileStore := &aerospike.AerospikeStore[api.UserProfile]{
		Client:    asClient,
		Policy:    as.NewPolicy(),
		Namespace: "mimuw",
		Set:       "user_profile",
		Bin:       "data",
	}
	//
	//gokaConfig := goka.DefaultConfig()
	//
	//tmc := goka.NewTopicManagerConfig()
	//tmc.Table.Replication = 1
	//tmc.Stream.Replication = 1
	//tmc.Stream.Retention = 24 * time.Hour
	//tmc.MismatchBehavior = goka.TMConfigMismatchBehaviorFail
	//
	//tm, err := goka.NewTopicManager(bootstrap, gokaConfig, tmc)
	//if err != nil {
	//	klog.Fatalf("can't create new goka topic manager: %v", err)
	//}
	//
	//err = tm.EnsureStreamExists(string(kafka.UserProfileTopic), 1)
	//if err != nil {
	//	klog.Fatalf("can't ensure stream \"%s\" exists: %v", kafka.UserProfileTopic, err)
	//}
	//
	//err = tm.EnsureTableExists(string(kafka.SinkTable), 1)
	//if err != nil {
	//	klog.Fatalf("can't ensure stream \"%s\" exists: %v", kafka.SinkTable, err)
	//}

	emitter, err := goka.NewEmitter(
		bootstrap,
		kafka.UserProfileTopic,
		new(codec.Bytes),
		//goka.WithEmitterTopicManagerBuilder(goka.TopicManagerBuilderWithTopicManagerConfig(tmc)),
		//goka.WithEmitterProducerBuilder(goka.ProducerBuilderWithConfig(gokaConfig)),
	)
	if err != nil {
		klog.Fatalf("can't create emitter: %v", err)
	}
	defer func() {
		err := emitter.Finish()
		if err != nil {
			klog.Errorf("can't finish emitter: %v", err)
		}
	}()

	view, err := goka.NewView(
		bootstrap,
		kafka.SinkTable,
		new(api.UserAggregatesCodec),
		//goka.WithViewTopicManagerBuilder(goka.TopicManagerBuilderWithTopicManagerConfig(tmc)),
		//goka.WithViewConsumerSaramaBuilder(goka.SaramaConsumerBuilderWithConfig(gokaConfig)),
	)
	if err != nil {
		klog.Fatalf("can't create view: %v", err)
	}

	var wg sync.WaitGroup
	defer wg.Wait()

	wg.Add(1)
	go func() {
		defer wg.Done()

		err = view.Run(ctx)
		if err != nil {
			close(stopCh)
			klog.Fatalf("can't run view: %v", err)
		}
	}()

	srv := server.NewHTTPServer(":8080", userProfileStore, emitter, view)

	wg.Add(1)
	go func() {
		defer wg.Done()

		klog.Infof("Starting web server on: %s", srv.Addr)

		err = srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			klog.Fatalf("Couldn't listen on %s: %v", srv.Addr, err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()

		klog.Info("Shutting down the server")
		defer log.Println("Server shut down")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			klog.Fatalf("Couldn't terminate gracefully: %v", err)
		}
	}()
}
