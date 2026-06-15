# Bililive Helper Go

> B 站 / 抖音直播录制后处理工具 — 自动合并、智能清理、Web 管理控制台

[![Go](https://img.shields.io/badge/Go-1.26-blue.svg)](https://golang.org/)
[![Vue](https://img.shields.io/badge/Vue-3.5-green.svg)](https://vuejs.org/)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

## 简介

自动监控 [bililive-go](https://github.com/hr3lxpesr6/bililive-go) 录制目录，将 FLV/TS 分片合并为完整视频，磁盘不足时智能清理旧文件，提供 Notion 风格 Web 控制台实时管理。

| 层 | 技术 |
|----|------|
| 后端 | Go 1.26、Gin 1.12、Zap |
| 前端 | Vue 3.5、TypeScript 6、Element Plus 2.14、Vite 8 |
| 视频 | FFmpeg（stream-copy 无损 + h264_rkmpp 硬件加速） |
| 部署 | Docker、Docker Compose（Alpine 多阶段构建） |

## 功能

- **录制合并**：FLV→MP4 无损转码、TS 拼接、硬件编码 fallback、文件健康预检、原子写入三重校验
- **智能清理**：磁盘阈值触发、白名单保护、安全期保护、紧急清理
- **任务调度**：事件驱动（目录变化检测）+ 固定间隔兜底、任务队列并发控制、静默时段
- **Web 控制台**：系统概览、主播管理、任务中心（SSE 实时输出）、操作日志、全局设置、智能推荐
- **安全设计**：bcrypt 哈希、登录限流、Session 版本化、CSP 安全头、路径穿越防护

## 快速开始

```bash
# 1. 克隆项目
git clone https://github.com/wuvjie/bililive-helper-go.git
cd bililive-helper-go

# 2. 修改 docker-compose.yml 中的录像目录（必须）

# 3. 构建镜像（首次需要，约 3-5 分钟）
docker compose build

# 4. 启动
docker compose up -d

# 5. 访问 Web 控制台（首次运行进入初始化向导）
open http://localhost:5689
```

> 首次启动自动进入初始化向导，设置登录密码即可。

### 海外构建

```bash
docker compose build \
  --build-arg ALPINE_MIRROR=dl-cdn.alpinelinux.org \
  --build-arg GO_PROXY=https://proxy.golang.org,direct \
  --build-arg NPM_REGISTRY=https://registry.npmjs.org
```

## 环境依赖

| 依赖 | 版本 | 说明 |
|------|------|------|
| Go | 1.26+ | 后端编译 |
| Node.js | 20.19+ | 前端编译（Vite 8 要求） |
| FFmpeg | 任意 | 视频转码、校验 |
| Docker | 20.10+ | 部署（推荐） |

## 开发

```bash
# 后端
go run ./cmd/server

# 前端（开发模式，API 代理到 :5000）
cd frontend && npm install && npm run dev

# 测试
go test ./...           # 后端
cd frontend && npm test # 前端
cd frontend && npm run lint # 代码检查
```

## 文档

- [API 接口文档](docs/api.md) — 所有 REST 端点说明
- [配置参考](docs/config.md) — 环境变量和 config.json 字段

## 常见问题

**ffmpeg 未安装** → `apt install ffmpeg` 或 Dockerfile 已内置

**忘记密码** → 删除 `{LOG_DIR}/.credentials.json` 后重启，自动生成新密码

**磁盘超阈值未触发清理** → 检查 `clean_enabled` 是否开启、是否在静默时段、文件是否在安全期内

**合并失败 "磁盘空间不足"** → TS 合并管线峰值需约 3 倍源文件大小空间

## 许可证

[MIT License](LICENSE)
