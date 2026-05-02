FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY bot/ ./
RUN go mod download && go build -o bot .

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/bot .

CMD ["./bot"]