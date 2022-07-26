all: push-image-service push-image-forwarder push-image-collector

SHELL :=/bin/bash -euEo pipefail -O inherit_errexit

SERVICE_IMAGE_TAG ?= latest
SERVICE_IMAGE_REF ?= docker.io/rzetelskik/allezon-service:$(SERVICE_IMAGE_TAG)

FORWARDER_IMAGE_TAG ?= latest
FORWARDER_IMAGE_REF ?= docker.io/rzetelskik/allezon-forwarder:$(FORWARDER_IMAGE_TAG)

COLLECTOR_IMAGE_TAG ?= latest
COLLECTOR_IMAGE_REF ?= docker.io/rzetelskik/allezon-collector:$(COLLECTOR_IMAGE_TAG)

image-service:
	docker build -f service/Dockerfile -t $(SERVICE_IMAGE_REF) .
.PHONY: image-service

image-forwarder:
	docker build -f forwarder/Dockerfile -t $(FORWARDER_IMAGE_REF) .
.PHONY: image-forwarder

image-collector:
	docker build -f collector/Dockerfile -t $(COLLECTOR_IMAGE_REF) .
.PHONY: image-collector

push-image-service: image-service
	docker push $(SERVICE_IMAGE_REF)
.PHONY: push-image-service

push-image-forwarder: image-forwarder
	docker push $(FORWARDER_IMAGE_REF)
.PHONY: push-image-forwarder

push-image-collector: image-collector
	docker push $(COLLECTOR_IMAGE_REF)
.PHONY: push-image-collector


