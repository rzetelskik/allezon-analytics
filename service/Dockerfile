FROM golang:1.18 as build

WORKDIR /go/src/github.com/rzetelskik/allezon-analytics/
COPY shared/ shared/
COPY service/go.mod service/go.sum service/

WORKDIR /go/src/github.com/rzetelskik/allezon-analytics/service/
RUN go mod download

COPY service/ ./
RUN make build

FROM ubuntu:21.04

COPY --from=build /go/src/github.com/rzetelskik/allezon-analytics/service/service /usr/bin/
RUN ln -s /usr/bin/service /service
EXPOSE 8080

ENTRYPOINT ["/usr/bin/service"]