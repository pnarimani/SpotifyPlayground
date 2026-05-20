FROM golang:1.26-alpine AS builder
WORKDIR /app

COPY contracts/go.mod ./contracts/
COPY playback/go.mod playback/go.sum ./playback/

RUN cd /app/playback && go mod download

COPY contracts ./contracts
COPY playback ./playback

WORKDIR /app/playback
RUN GOWORK=off CGO_ENABLED=0 GOOS=linux go build -o /app/server .

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/server ./server
CMD ["./server"]
