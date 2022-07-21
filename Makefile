all: image-service image-forwarder image-collector

SHELL :=/bin/bash -euEo pipefail -O inherit_errexit

SERVICE_IMAGE_TAG ?= latest
SERVICE_IMAGE_REF ?= docker.io/rzetelskik/allezon-analytics-service:$(SERVICE_IMAGE_TAG)

FORWARDER_IMAGE_TAG ?= latest
FORWARDER_IMAGE_REF ?= docker.io/rzetelskik/allezon-analytics-forwarder:$(FORWARDER_IMAGE_TAG)

COLLECTOR_IMAGE_TAG ?= latest
COLLECTOR_IMAGE_REF ?= docker.io/rzetelskik/allezon-analytics-collector:$(COLLECTOR_IMAGE_TAG)

image-service:
	docker build -f service/Dockerfile -t $(SERVICE_IMAGE_REF) .
.PHONY: image-service

image-forwarder:
	docker build -f forwarder/Dockerfile -t $(FORWARDER_IMAGE_REF) .
.PHONY: image-forwarder

image-collector:
	docker build -f collector/Dockerfile -t $(COLLECTOR_IMAGE_REF) .
.PHONY: image-collector