FROM golang:1.25.1-alpine AS builder-server

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -o server ./cmd/api/

FROM golang:1.25.1-alpine AS builder-migrator

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -o migrator ./cmd/migrator/

FROM alpine:latest AS api

WORKDIR /app
COPY --from=builder-server /build/server .
CMD ["./server"]

FROM alpine:latest AS migrator

WORKDIR /app
COPY --from=builder-migrator /build/migrator .
COPY --from=builder-migrator /build/migrations/ ./migrations/
CMD ["./migrator"]
