# Build stage
FROM golang:1.18-alpine3.15 AS builder
WORKDIR /app
COPY . .
ENV GOPROXY https://goproxy.cn
RUN go build -o main main.go

# Run stage
FROM alpine
WORKDIR /app
COPY --from=builder /app/main .
COPY config.env .

EXPOSE 8080
CMD ["/app/main"]