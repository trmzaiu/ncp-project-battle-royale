# Build stage
FROM golang:1.24 AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server main.go

# Run stage
FROM gcr.io/distroless/base-debian11
WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/assets ./assets

CMD ["./server"]
EXPOSE 8080
