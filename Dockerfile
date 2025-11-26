# build stage
FROM golang:1.25.4-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
RUN go install github.com/swaggo/swag/cmd/swag@v1.16.6
COPY . .
RUN swag init -g cmd/server/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/server

# run stage
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/.env . 
EXPOSE 8080
CMD ["./main"]