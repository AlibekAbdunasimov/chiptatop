# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS builder
WORKDIR /app

COPY go.mod ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bot ./cmd/bot

FROM alpine:3.20
RUN apk add --no-cache ca-certificates && adduser -D -g '' appuser
USER appuser
WORKDIR /home/appuser
COPY --from=builder /app/bot /usr/local/bin/bot
ENV ENVIRONMENT=production
ENTRYPOINT ["/usr/local/bin/bot"]


