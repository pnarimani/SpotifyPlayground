FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server

FROM alpine
WORKDIR /app
COPY --from=builder /app/server .
CMD [ "./server" ]