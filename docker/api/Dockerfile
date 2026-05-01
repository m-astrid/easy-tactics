FROM golang:1.21-alpine

WORKDIR /app

COPY api/ ./
COPY migrations/ ./migrations/
RUN go mod download && go build -o server .

EXPOSE 50051

CMD ["./server"]