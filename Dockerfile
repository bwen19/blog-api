# Build stage
FROM golang:1.19-alpine3.16 AS builder
WORKDIR /app
COPY . .
ENV GOPROXY https://goproxy.cn
RUN go build -o main main.go

# Run stage
FROM alpine:3.16
WORKDIR /app
COPY --from=builder /app/main .
COPY config.env .
COPY db/migration ./db/migration

EXPOSE 8080
CMD ["/app/main"]