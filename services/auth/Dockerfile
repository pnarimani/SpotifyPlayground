FROM golang:1.26-alpine AS builder
WORKDIR /app

COPY middleware/go.mod ./middleware/
COPY auth/go.mod auth/go.sum ./auth/

RUN cd /app/auth && go mod download

COPY middleware ./middleware
COPY auth ./auth

WORKDIR /app/auth
RUN GOWORK=off CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/server ./server
CMD ["./server"]
