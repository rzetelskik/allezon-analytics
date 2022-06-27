all: build

SHELL :=/bin/bash -euEo pipefail -O inherit_errexit

IMAGE_TAG ?= latest
IMAGE_REF ?= docker.io/rzetelskik/allezon-analytics:$(IMAGE_TAG)

build:
	CGO_ENABLED=0 GOOS=linux go build github.com/rzetelskik/allezon-analytics/cmd/allezon-analytics

image:
	docker build . -t $(IMAGE_REF)
.PHONY: image