FROM golang:1.24-bullseye AS builder

WORKDIR /app

COPY . .

RUN go mod download

RUN go build -o protocache ./cmd/protocache
RUN go build -o protocachecli ./cmd/protocachecli

FROM debian:bullseye-slim

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/protocache /usr/local/bin/protocache
COPY --from=builder /app/protocachecli /usr/local/bin/protocachecli

ENTRYPOINT ["/usr/local/bin/protocache"]
