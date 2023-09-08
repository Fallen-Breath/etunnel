FROM golang:1.20 as builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./
RUN CGO_ENABLED=0 go build github.com/Fallen-Breath/etunnel/cmd/etunnel

FROM alpine:latest
WORKDIR /root
COPY --from=builder /build/etunnel /usr/local/bin/etunnel

ENTRYPOINT ["/usr/local/bin/etunnel"]
CMD ["--help"]
