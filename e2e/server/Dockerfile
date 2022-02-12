FROM golang:1.17.6-alpine3.15 AS builder

# Install packages
RUN apk add --no-cache bash gcc musl-dev linux-headers libpcap-dev

# Compile application
WORKDIR /go/src/github.com/guillembonet/go-wireguard-udpholepunch
ADD . .
RUN go build -o build/server e2e/server/server.go

FROM alpine:3.15

# Install packages
RUN apk add --no-cache ca-certificates tcpdump sudo libpcap-dev

# Install application
COPY --from=builder /go/src/github.com/guillembonet/go-wireguard-udpholepunch/build/server /usr/bin/server

ENTRYPOINT ["/usr/bin/server"]