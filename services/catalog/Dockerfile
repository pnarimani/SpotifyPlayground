FROM golang:1.26-alpine AS builder
WORKDIR /app

COPY middleware/go.mod ./middleware/
COPY catalog/go.mod catalog/go.sum ./catalog/

RUN cd /app/catalog && go mod download

COPY middleware ./middleware
COPY catalog ./catalog

WORKDIR /app/catalog
RUN GOWORK=off CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/server ./server
CMD ["./server"]
