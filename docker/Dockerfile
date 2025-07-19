FROM golang:1.24 AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./
RUN CGO_ENABLED=0 go build github.com/Fallen-Breath/etunnel/cmd/etunnel

FROM alpine:latest
RUN apk add --no-cache tzdata
WORKDIR /root
COPY --from=builder /build/etunnel /usr/local/bin/etunnel

ENTRYPOINT ["/usr/local/bin/etunnel"]
CMD ["--help"]
