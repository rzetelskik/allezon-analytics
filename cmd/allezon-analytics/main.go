package main

import (
	"context"
	"errors"
	"flag"
	as "github.com/aerospike/aerospike-client-go/v6"
	"github.com/rzetelskik/allezon-analytics/internal/allezon-analytics/aerospike"
	"github.com/rzetelskik/allezon-analytics/internal/allezon-analytics/api"
	"github.com/rzetelskik/allezon-analytics/internal/allezon-analytics/server"
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

	srv := server.NewHTTPServer(":8080", userProfileStore)

	var wg sync.WaitGroup
	defer wg.Wait()

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
