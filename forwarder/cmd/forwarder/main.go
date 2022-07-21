package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"github.com/lovoo/goka"
	"github.com/lovoo/goka/codec"
	"github.com/rzetelskik/allezon-analytics/shared/pkg/api"
	"k8s.io/klog/v2"
	"log"
	"os"
	"runtime"
	"time"
)

var (
	bootstrap                  = []string{"kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"}
	group          goka.Group  = "forwarder"
	topic          goka.Stream = "kafka-topic"
	aggregateTopic goka.Stream = "aggregate-topic"
)

func backtrack(pos int, curr string, ss []string, res *[]string) {
	if pos == len(ss) {
		*res = append(*res, curr)
		return
	}

	for i := pos; i < len(ss); i++ {
		backtrack(i+1, curr+ss[i], ss, res)
	}
}

func forward(ctx goka.Context, i interface{}) {
	ut, ok := i.(*api.UserTag)
	if !ok {
		panic("aaaaa")
	}

	bucket := ut.Time.Truncate(time.Minute)

	base := bucket.String() + ut.Action.String()
	ss := []string{ut.Origin, ut.Product.BrandID, ut.Product.CategoryID}

	res := []string{base}
	for i := 0; i < len(ss); i++ {
		backtrack(i, base, ss, &res)
	}

	for _, key := range res {
		klog.InfoS("key", key)
		h := sha256.New()
		h.Write([]byte(key))
		hash := hex.EncodeToString(h.Sum(nil))
		ctx.Emit(aggregateTopic, hash, int64(ut.Product.Price))
	}
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

	err = tm.EnsureStreamExists(string(topic), 1)
	if err != nil {
		klog.Fatalf("can't ensure stream exists: %v", err)
	}

	err = tm.EnsureStreamExists(string(aggregateTopic), 1)
	if err != nil {
		klog.Fatalf("can't ensure stream exists: %v", err)
	}

	g := goka.DefineGroup(group,
		goka.Input(topic, new(api.UserTagCodec), forward),
		goka.Output(aggregateTopic, new(codec.Int64)),
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
