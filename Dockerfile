# 镜像源配置（海外构建传 --build-arg ALPINE_MIRROR=dl-cdn.alpinelinux.org）
ARG ALPINE_MIRROR=mirrors.aliyun.com
ARG GO_PROXY=https://goproxy.cn,direct
ARG NPM_REGISTRY=https://registry.npmmirror.com

# Stage 1: Build frontend
FROM node:20-alpine AS frontend-builder
ARG NPM_REGISTRY
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm config set registry ${NPM_REGISTRY} && npm ci
COPY frontend/ ./
RUN npm run build

# Stage 2: Build Go binary
FROM golang:1.26-alpine AS builder
ARG ALPINE_MIRROR
ARG GO_PROXY

RUN sed -i "s/dl-cdn.alpinelinux.org/${ALPINE_MIRROR}/g" /etc/apk/repositories
RUN apk add --no-cache git
ENV GOPROXY=${GO_PROXY}

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o bililive-helper-go ./cmd/server

# Stage 3: Minimal runtime image with ffmpeg
FROM alpine:3.19
ARG ALPINE_MIRROR

LABEL maintainer="wuvjie"
LABEL description="Bililive Helper Go - 直播录制后处理工具"
LABEL org.opencontainers.image.source="https://github.com/wuvjie/bililive-helper-go"

RUN sed -i "s/dl-cdn.alpinelinux.org/${ALPINE_MIRROR}/g" /etc/apk/repositories
RUN apk add --no-cache ffmpeg ca-certificates tzdata curl

WORKDIR /app

COPY --from=builder /app/bililive-helper-go .
COPY --from=frontend-builder /app/templates ./templates
# login.html is a server-side template not built by Vite — copy from source
COPY --from=builder /app/templates/login.html ./templates/login.html

EXPOSE 5000
CMD ["./bililive-helper-go"]
