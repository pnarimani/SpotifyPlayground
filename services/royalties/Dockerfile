FROM golang:1.26-alpine AS builder
WORKDIR /app

COPY contracts/go.mod ./contracts/
COPY royalties/go.mod royalties/go.sum ./royalties/

RUN cd /app/royalties && go mod download

COPY contracts ./contracts
COPY royalties ./royalties

WORKDIR /app/royalties
RUN GOWORK=off CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/server ./server
CMD ["./server"]
