all: build

build:
	CGO_ENABLED=0 GOOS=linux go build github.com/rzetelskik/allezon-analytics/cmd/allezon-analytics