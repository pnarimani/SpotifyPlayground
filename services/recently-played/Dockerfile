FROM golang:1.26-alpine AS builder
WORKDIR /app

COPY contracts/go.mod ./contracts/
COPY middleware/go.mod middleware/go.sum ./middleware/
COPY recently-played/go.mod recently-played/go.sum ./recently-played/

RUN cd /app/middleware && go mod download && go mod verify \
 && cd /app/recently-played && go mod download

COPY contracts ./contracts
COPY middleware ./middleware
COPY recently-played ./recently-played

WORKDIR /app/recently-played
RUN GOWORK=off CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/server ./server
CMD ["./server"]
