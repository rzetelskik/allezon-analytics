FROM golang:1.18 as build

WORKDIR /go/src/github.com/rzetelskik/allezon-analytics/
COPY shared/ shared/

COPY collector/ collector/

WORKDIR /go/src/github.com/rzetelskik/allezon-analytics/collector/
RUN make build

FROM ubuntu:21.04

COPY --from=build /go/src/github.com/rzetelskik/allezon-analytics/collector/collector /usr/bin/
RUN ln -s /usr/bin/collector /collector
EXPOSE 8080

ENTRYPOINT ["/usr/bin/collector"]