FROM golang:1.25.3-alpine3.22 AS builder

RUN apk add --update ca-certificates git

ENV SRC github.com/FuelLabs/segment-proxy
ENV CGO_ENABLED=0
ENV GO111MODULE=on
ENV GOOS=linux

ARG VERSION

COPY . /go/src/${SRC}
WORKDIR /go/src/${SRC}

RUn go build -a -installsuffix cgo -ldflags "-w -s -extldflags '-static' -X main.version=$VERSION" -o /proxy

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /proxy /proxy

EXPOSE 8080

ENTRYPOINT ["/proxy"]
