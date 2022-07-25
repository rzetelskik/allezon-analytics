package forwarder

import (
	"github.com/lovoo/goka"
	"github.com/rzetelskik/allezon-analytics/shared/pkg/api"
	"github.com/rzetelskik/allezon-analytics/shared/pkg/kafka"
	"github.com/rzetelskik/allezon-analytics/shared/pkg/util"
	"k8s.io/klog/v2"
	"time"
)

func backtrack(pos int, curr []string, ss []string, res *[][]string) {
	if pos == len(ss) {
		*res = append(*res, curr)
		return
	}

	for i := pos; i < len(ss); i++ {
		backtrack(i+1, append(curr, ss[i]), ss, res)
	}
}

func Forward(ctx goka.Context, i interface{}) {
	ut, ok := i.(*api.UserTag)
	if !ok {
		klog.Fatalf("received interface is not a UserTag")
	}

	properties := []string{ut.Origin, ut.Product.BrandID, ut.Product.CategoryID}
	filters := make([][]string, 0)
	for i := 0; i < len(properties); i++ {
		backtrack(i, []string{}, properties, &filters)
	}

	bucket := ut.Time.Truncate(time.Minute)
	for _, f := range filters {
		hash := util.GetAggregateHash(bucket, ut.Action, f...)
		ctx.Emit(kafka.AggregateTopic, hash, int64(ut.Product.Price))
	}
}
