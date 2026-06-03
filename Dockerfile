# Stage 1: Build frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# Stage 2: Build Go binary
FROM golang:1.25-alpine AS builder

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN apk add --no-cache git
ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o bililive-helper ./cmd/server

# Stage 3: Minimal runtime image with ffmpeg
FROM alpine:3.19

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN apk add --no-cache ffmpeg ca-certificates tzdata curl

WORKDIR /app

COPY --from=builder /app/bililive-helper .
COPY --from=frontend-builder /app/templates ./templates
# login.html is a server-side template not built by Vite — copy from source
COPY --from=builder /app/templates/login.html ./templates/login.html

RUN mkdir -p /data

# Run as non-root user for security
RUN adduser -D -u 1000 appuser
USER appuser

EXPOSE 5000
CMD ["./bililive-helper"]
