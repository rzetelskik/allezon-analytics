FROM golang:1.18 as build

WORKDIR /go/src/github.com/rzetelskik/allezon-analytics/
COPY shared/ shared/

COPY forwarder/ forwarder/

WORKDIR /go/src/github.com/rzetelskik/allezon-analytics/forwarder/
RUN make build

FROM ubuntu:21.04

COPY --from=build /go/src/github.com/rzetelskik/allezon-analytics/forwarder/forwarder /usr/bin/
RUN ln -s /usr/bin/forwarder /forwarder
EXPOSE 8080

ENTRYPOINT ["/usr/bin/forwarder"]