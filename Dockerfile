# 镜像源配置（默认使用官方通用标准源，国内构建可通过 --build-arg 传入加速镜像源）
ARG REGISTRY=docker.io
ARG ALPINE_MIRROR=dl-cdn.alpinelinux.org
ARG GO_PROXY=https://proxy.golang.org,direct
ARG NPM_REGISTRY=https://registry.npmjs.org

# Stage 1: Build frontend
FROM --platform=$BUILDPLATFORM ${REGISTRY}/library/node:20-alpine AS frontend-builder
ARG NPM_REGISTRY
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm config set registry ${NPM_REGISTRY} && npm ci
COPY frontend/ ./
RUN npm run build

# Stage 2: Build Go binary (使用宿主机原生架构通过 GOARCH 极速交叉编译)
FROM --platform=$BUILDPLATFORM ${REGISTRY}/library/golang:1.25-alpine AS builder
ARG ALPINE_MIRROR
ARG GO_PROXY
ARG TARGETOS=linux
ARG TARGETARCH=amd64

RUN sed -i "s/dl-cdn.alpinelinux.org/${ALPINE_MIRROR}/g" /etc/apk/repositories
RUN apk add --no-cache git
ENV GOPROXY=${GO_PROXY}

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-s -w" -o bililive-helper-go ./cmd/server

# Stage 3: Minimal runtime image with ffmpeg (针对目标架构生成精简运行时)
FROM ${REGISTRY}/library/alpine:3.21
ARG ALPINE_MIRROR

LABEL maintainer="The Bililive Helper Go Authors"
LABEL description="Bililive Helper Go - 直播录制后处理工具"
LABEL org.opencontainers.image.source="https://github.com/wuvjie/bililive-helper-go"

RUN sed -i "s/dl-cdn.alpinelinux.org/${ALPINE_MIRROR}/g" /etc/apk/repositories
RUN apk add --no-cache ffmpeg ca-certificates tzdata curl

# 默认设置时区为中国标准时间（Asia/Shanghai）
ENV TZ=Asia/Shanghai

WORKDIR /app

COPY --from=builder /app/bililive-helper-go .
COPY --from=frontend-builder /app/templates ./templates
# login.html is a server-side template not built by Vite — copy from source
COPY --from=builder /app/templates/login.html ./templates/login.html

EXPOSE 5000
CMD ["./bililive-helper-go"]
