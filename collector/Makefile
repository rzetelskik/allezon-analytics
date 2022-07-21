all: build

SHELL :=/bin/bash -euEo pipefail -O inherit_errexit

build:
	CGO_ENABLED=0 GOOS=linux go build github.com/rzetelskik/allezon-analytics/collector/cmd/collector
