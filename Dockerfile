# syntax=docker/dockerfile:1

FROM golang:1.25.1 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/server

FROM debian:bookworm-slim
WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/server /app/server
COPY --from=builder /app/docs /app/docs

ENV DATABASE_URL=postgres://in:in@db:5432/lindb
EXPOSE 8080

CMD ["/app/server"]
