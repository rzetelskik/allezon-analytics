package kafka

import "github.com/lovoo/goka"

const (
	UserProfileTopic goka.Stream = "user-profile"
	AggregateTopic   goka.Stream = "aggregate"
	SinkGroup        goka.Group  = "collector"
	SinkTable        goka.Table  = "collector-table"
)
