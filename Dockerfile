FROM golang:1.23-alpine AS builder

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN apk add --no-cache git
ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o bililive-helper ./cmd/server

FROM alpine:latest

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN apk add --no-cache ffmpeg ca-certificates tzdata curl

WORKDIR /app

COPY --from=builder /app/bililive-helper .
COPY templates ./templates

RUN mkdir -p /data
EXPOSE 5000
CMD ["./bililive-helper"]
