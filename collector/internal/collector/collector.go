package collector

import (
	"github.com/lovoo/goka"
	"github.com/rzetelskik/allezon-analytics/shared/pkg/api"
	"k8s.io/klog/v2"
)

type Collector struct {
	store *Store
}

func (c *Collector) Collect(ctx goka.Context, msg interface{}) {
	//var ua api.UserAggregates

	//v := ctx.Value()
	//if v != nil {
	//	ua = v.(api.UserAggregates)
	//}
	key := ctx.Key()

	ut, ok := msg.(*api.UserTag)
	if !ok {
		klog.Errorf("received message's type is not of type UserTag")
		return
	}

	c.store.Add(key, api.UserAggregates{
		Count:    1,
		SumPrice: int64(ut.Product.Price),
	})

	//ua.Count += 1
	//ua.SumPrice += int64(ut.Product.Price)

	//ctx.SetValue(ua)
}

func NewCollector(store *Store) Collector {
	return Collector{
		store: store,
	}
}
