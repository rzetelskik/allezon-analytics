FROM golang:1.18 as build

WORKDIR /go/src/github.com/rzetelskik/allezon-analytics

COPY . .

RUN make build

FROM ubuntu:21.04

COPY --from=build /go/src/github.com/rzetelskik/allezon-analytics/allezon-analytics /usr/bin/

EXPOSE 8080

ENTRYPOINT ["/usr/bin/allezon-analytics"]