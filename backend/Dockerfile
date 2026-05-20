# syntax=docker/dockerfile:1

FROM golang:1.23-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git ca-certificates

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/ingest ./cmd/ingest
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/migrate ./cmd/migrate

FROM alpine:3.20 AS api
RUN apk add --no-cache ca-certificates
COPY --from=builder /bin/api /usr/local/bin/api
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/api"]

FROM alpine:3.20 AS ingest
RUN apk add --no-cache ca-certificates
COPY --from=builder /bin/ingest /usr/local/bin/ingest
ENTRYPOINT ["/usr/local/bin/ingest"]

FROM alpine:3.20 AS migrate
RUN apk add --no-cache ca-certificates
COPY --from=builder /bin/migrate /usr/local/bin/migrate
ENTRYPOINT ["/usr/local/bin/migrate"]
