package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/lovoo/goka"
	"github.com/lovoo/goka/codec"
	"k8s.io/klog/v2"
	"log"
	"os"
	"runtime"
)

var (
	bootstrap                  = []string{"kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"}
	group          goka.Group  = "collector"
	aggregateTopic goka.Stream = "aggregate-topic"
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

func collect(ctx goka.Context, msg interface{}) {
	var ua UserAggregates

	v := ctx.Value()
	if v != nil {
		ua = v.(UserAggregates)
	}

	price, ok := msg.(int64)
	if !ok {
		panic("aaa!")
	}

	ua.Count += 1
	ua.SumPrice += price

	ctx.SetValue(ua)
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

	config := goka.DefaultConfig()

	tmc := goka.NewTopicManagerConfig()
	tmc.Table.Replication = 1
	tmc.Stream.Replication = 1

	tm, err := goka.NewTopicManager(bootstrap, config, tmc)
	if err != nil {
		klog.Fatalf("can't create topic manager: %v", err)
	}
	defer tm.Close()

	err = tm.EnsureStreamExists(string(aggregateTopic), 1)
	if err != nil {
		klog.Fatalf("can't ensure stream exists: %v", err)
	}

	g := goka.DefineGroup(group,
		goka.Input(aggregateTopic, new(codec.Int64), collect),
		goka.Persist(new(UserAggregatesCodec)),
	)

	p, err := goka.NewProcessor(bootstrap,
		g,
		goka.WithTopicManagerBuilder(goka.TopicManagerBuilderWithTopicManagerConfig(tmc)),
		goka.WithConsumerGroupBuilder(goka.ConsumerGroupBuilderWithConfig(config)),
	)
	if err != nil {
		klog.Fatalf("can't create new processor: %v", err)
	}

	err = p.Run(context.Background())
	if err != nil {
		klog.Fatalf("can't run processor: %v", err)
	}
}
