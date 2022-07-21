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

	ut, ok := msg.(*api.UserTag)
	if !ok {
		klog.Errorf("received message's type is not of type UserTag")
		return
	}

	ua.Count += 1
	ua.SumPrice += int64(ut.Product.Price)

	ctx.SetValue(ua)
}
