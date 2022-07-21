module github.com/rzetelskik/allezon-analytics/service

go 1.18

require (
	github.com/aerospike/aerospike-client-go/v6 v6.2.0
	github.com/golang/snappy v0.0.4
	github.com/google/go-cmp v0.5.7
	github.com/gorilla/mux v1.8.0
	github.com/lovoo/goka v1.1.7
	github.com/rzetelskik/allezon-analytics/shared v0.0.0
	k8s.io/klog/v2 v2.70.0
)

require github.com/Shopify/sarama v1.33.0 // indirect

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/eapache/go-resiliency v1.2.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20180814174437-776d5712da21 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/go-logr/logr v1.2.0 // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/golang/mock v1.4.4 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-uuid v1.0.2 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.0.0 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.2 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/klauspost/compress v1.15.6 // indirect
	github.com/onsi/ginkgo v1.16.5 // indirect
	github.com/onsi/gomega v1.19.0 // indirect
	github.com/pierrec/lz4 v2.6.1+incompatible // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/syndtr/goleveldb v1.0.0 // indirect
	github.com/yuin/gopher-lua v0.0.0-20220504180219-658193537a64 // indirect
	golang.org/x/crypto v0.0.0-20220214200702-86341886e292 // indirect
	golang.org/x/net v0.0.0-20220520000938-2e3eb7b945c2 // indirect
	golang.org/x/sync v0.0.0-20220601150217-0de741cfad7f // indirect
)

replace github.com/rzetelskik/allezon-analytics/shared v0.0.0 => ../shared
