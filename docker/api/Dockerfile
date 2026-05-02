FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY api/ ./
RUN go mod download && go build -o server .

FROM alpine:latest
RUN apk add --no-cache sqlite-client
RUN go install github.com/pressly/goose/v3@latest

WORKDIR /app
COPY --from=builder /app/server .
COPY migrations/ ./migrations/
COPY api/proto/gen/ ./proto/gen/

EXPOSE 50051

CMD ["sh", "-c", "goose -dir migrations sqlite3 /data/fighters.db up && ./server"]