package main

import (
	"context"
	"flag"
	"github.com/lovoo/goka"
	"github.com/rzetelskik/allezon-analytics/collector/internal/collector"
	"github.com/rzetelskik/allezon-analytics/shared/pkg/api"
	"github.com/rzetelskik/allezon-analytics/shared/pkg/kafka"
	"k8s.io/klog/v2"
	"log"
	"os"
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
	log.SetOutput(os.Stdout)

	g := goka.DefineGroup(kafka.SinkGroup,
		goka.Input(kafka.AggregateTopic, new(api.UserTagCodec), collector.Collect),
		goka.Persist(new(api.UserAggregatesCodec)),
	)

	p, err := goka.NewProcessor(
		[]string{kafka.Bootstrap},
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
