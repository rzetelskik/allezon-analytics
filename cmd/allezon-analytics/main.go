package main

import (
	"flag"
	as "github.com/aerospike/aerospike-client-go"
	"github.com/rzetelskik/allezon-analytics/internal/allezon-analytics/aerospike"
	"github.com/rzetelskik/allezon-analytics/internal/allezon-analytics/api"
	"github.com/rzetelskik/allezon-analytics/internal/allezon-analytics/server"
	"k8s.io/klog/v2"
	"runtime"
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

	host := as.NewHost("st111vm102.rtb-lab.pl", 3000)
	policy := as.NewClientPolicy()

	asClient, err := as.NewClientWithPolicyAndHost(policy, host)
	if err != nil {
		klog.Fatalf("can't create new Aerospike client: %v", err)
	}
	defer asClient.Close()

	userProfileStore := &aerospike.AerospikeStore[api.UserProfile]{
		Client:    asClient,
		Policy:    as.NewPolicy(),
		Namespace: "test",
		Set:       "user_profile",
		Bin:       "data",
	}

	srv := server.NewHTTPServer(":8080", userProfileStore)

	klog.Info("Starting web server...")
	klog.Fatal(srv.ListenAndServe())
}
