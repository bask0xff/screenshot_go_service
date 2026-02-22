FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN go build -o server .

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8082

CMD ["./server"]