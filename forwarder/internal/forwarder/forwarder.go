package forwarder

import (
	"github.com/lovoo/goka"
	"github.com/rzetelskik/allezon-analytics/shared/pkg/api"
	"github.com/rzetelskik/allezon-analytics/shared/pkg/kafka"
	"github.com/rzetelskik/allezon-analytics/shared/pkg/util"
	"k8s.io/klog/v2"
	"time"
)

func Forward(ctx goka.Context, msg interface{}) {
	ut, ok := msg.(*api.UserTag)
	if !ok {
		klog.Errorf("received message's type is not of type UserTag")
		return
	}

	properties := []string{ut.Origin, ut.Product.BrandID, ut.Product.CategoryID}
	filters := make([][]string, 0)
	for i := 0; i < len(properties); i++ {
		Backtrack(i, []string{}, properties, &filters)
	}

	bucket := ut.Time.Truncate(time.Minute)
	for _, f := range filters {
		hash := util.GetAggregateHash(bucket, ut.Action, f...)
		ctx.Emit(kafka.AggregateTopic, hash, ut)
	}
}
