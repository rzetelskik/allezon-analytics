package main

import (
	"context"
	"flag"
	"github.com/lovoo/goka"
	"github.com/lovoo/goka/codec"
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

	//err = kafkautil.EnsureStreamExists(kafka.UserProfileTopic, 1, bootstrap)
	//if err != nil {
	//	klog.Fatalf("can't ensure stream \"%s\" exists: %v", kafka.UserProfileTopic, err)
	//}
	//
	//err = kafkautil.EnsureStreamExists(kafka.AggregateTopic, 1, bootstrap)
	//if err != nil {
	//	klog.Fatalf("can't ensure stream \"%s\" exists: %v", kafka.AggregateTopic, err)
	//
	//}

	var group goka.Group = "forwarder"
	g := goka.DefineGroup(group,
		goka.Input(kafka.UserProfileTopic, new(api.UserTagCodec), forwarder.Forward),
		goka.Output(kafka.AggregateTopic, new(codec.Int64)),
	)

	p, err := goka.NewProcessor(bootstrap, g)
	if err != nil {
		klog.Fatalf("can't create new processor: %v", err)
	}

	err = p.Run(context.Background())
	if err != nil {
		klog.Fatalf("can't run processor: %v", err)
	}
}
