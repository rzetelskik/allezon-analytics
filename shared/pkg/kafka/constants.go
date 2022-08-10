package kafka

import "github.com/lovoo/goka"

const (
	Bootstrap        string      = "kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"
	UserProfileTopic goka.Stream = "user-profile"
	AggregateTopic   goka.Stream = "aggregate"
	SinkGroup        goka.Group  = "collector"
	SinkTable        goka.Table  = "collector-table"
)
