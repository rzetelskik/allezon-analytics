package collector

import (
	"github.com/lovoo/goka"
	"github.com/rzetelskik/allezon-analytics/shared/pkg/api"
	"k8s.io/klog/v2"
)

func Collect(ctx goka.Context, msg interface{}) {
	var ua api.UserAggregates

	v := ctx.Value()
	if v != nil {
		ua = v.(api.UserAggregates)
	}

	price, ok := msg.(int64)
	if !ok {
		klog.Fatalf("received message is not of type int64")
	}

	ua.Count += 1
	ua.SumPrice += price

	ctx.SetValue(ua)
}
