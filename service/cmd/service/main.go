package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	as "github.com/aerospike/aerospike-client-go/v6"
	"github.com/lovoo/goka"
	"github.com/lovoo/goka/codec"
	"github.com/rzetelskik/allezon-analytics/service/internal/service/aerospike"
	"github.com/rzetelskik/allezon-analytics/service/internal/service/server"
	"github.com/rzetelskik/allezon-analytics/shared/pkg/api"
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

type UserAggregates struct {
	Count    int64 `json:"count"`
	SumPrice int64 `json:"sum_price"`
}

type UserAggregatesCodec struct{}

func (c *UserAggregatesCodec) Encode(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (c *UserAggregatesCodec) Decode(data []byte) (interface{}, error) {
	var ua UserAggregates
	err := json.Unmarshal(data, &ua)
	if err != nil {
		return nil, fmt.Errorf("can't unmarshal data: %w", err)
	}

	return ua, nil
}

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

	var topic goka.Stream = "kafka-topic"
	emitter, err := goka.NewEmitter(bootstrap, topic, new(codec.Bytes))
	if err != nil {
		klog.Fatalf("can't create emitter: %v", err)
	}
	defer emitter.Finish()

	var group goka.Group = "collector"
	var table = goka.GroupTable(group)
	view, err := goka.NewView(bootstrap, table, new(UserAggregatesCodec))
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
