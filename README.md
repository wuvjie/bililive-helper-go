# Bililive Helper Go

> B 站 / 抖音直播录制后处理工具 — 自动合并、智能清理、Web 管理控制台

[![Go](https://img.shields.io/badge/Go-1.26-blue.svg)](https://golang.org/)
[![Vue](https://img.shields.io/badge/Vue-3.5-green.svg)](https://vuejs.org/)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

## 简介

自动监控 [bililive-go](https://github.com/hr3lxpesr6/bililive-go) 录制目录，将 FLV/TS 分片合并为完整视频，磁盘不足时智能清理旧文件，提供现代 Web 控制台实时管理。

| 层 | 技术 |
|----|------|
| 后端 | Go 1.26、Gin 1.12、Zap |
| 前端 | Vue 3.5、TypeScript 6、Element Plus 2.14、Vite 8 |
| 视频 | FFmpeg（stream-copy 无损 + h264_rkmpp 硬件加速） |
| 部署 | Docker、Docker Compose、WSL Container |

## 功能

- **录制合并**：FLV→MP4 无损转码、TS Concat Demuxer 稳定拼接、时间戳对齐、XML 弹幕与封面伴生对齐
- **智能清理**：磁盘阈值触发、白名单保护、安全期保护、一键免删、伴生弹幕与封面级联清理
- **任务调度**：事件驱动（目录变化检测）+ 固定间隔兜底、任务队列并发控制、静默时段
- **Web 控制台**：系统概览、主播管理、文件抽屉与在线视频预览、任务中心（SSE 实时日志）、操作审计、全局策略与备份
- **安全设计**：bcrypt 哈希、登录限流、Session 版本化、CSP 安全头、路径穿越防护

## 部署运行

### 1. 构建镜像（首次构建）

```bash
# 1. 克隆项目
git clone https://github.com/wuvjie/bililive-helper-go.git
cd bililive-helper-go

# 2. 构建镜像（约 2-4 分钟）
docker compose build

# 海外 VPS 或直连网络加速构建：
# docker compose build \
#   --build-arg ALPINE_MIRROR=dl-cdn.alpinelinux.org \
#   --build-arg GO_PROXY=https://proxy.golang.org,direct \
#   --build-arg NPM_REGISTRY=https://registry.npmjs.org
```

---

### 2. 启动服务（选择适合您环境的方式）

#### 方式一：Docker Compose（推荐）

1. **复制环境配置模板**：
   ```bash
   cp .env.example .env
   ```
2. **编辑 `.env` 文件**，填入宿主机上的实际路径：
   ```env
   # 宿主机上 bililive-go 的录制输出根目录
   RECORDINGS_DIR=/volume1/video/recordings

   # 宿主机上存放配置、凭据与日志的持久化目录
   DATA_DIR=/volume1/docker/bililive-helper-go

   # Web 访问端口（默认 5689）
   PORT=5689
   ```
3. **启动容器**：
   ```bash
   docker compose up -d
   ```

---

#### 方式二：Docker CLI 原生命令行（Linux / NAS / macOS）

使用标准 `docker run` 命令直接启动容器（将宿主机目录挂载至容器内部通用的 `/recordings` 与 `/data`）：

```bash
docker run -d \
  --name bililive-helper-go \
  --restart always \
  -p 5689:5000 \
  -v /path/to/recordings:/recordings \
  -v /path/to/data:/data \
  -e TARGET_DIR=/recordings \
  -e LOG_DIR=/data \
  -e TZ=Asia/Shanghai \
  --cpus="3.0" \
  --memory="4g" \
  --log-opt max-size=10m \
  --log-opt max-file=3 \
  bililive-helper-go
```

> **参数说明**：
> - `-v /path/to/recordings:/recordings`：将宿主机录像目录映射到容器 `/recordings`。
> - `-v /path/to/data:/data`：将宿主机配置和日志数据目录映射到容器 `/data`。
> - `-p 5689:5000`：宿主机对外暴露 5689 端口（容器内监听 5000）。

---

#### 方式三：WSL Container（Windows PowerShell + wslc）

在 Windows 环境下使用 WSL Container CLI (`wslc`) 时，可在 PowerShell 中使用反引号（`）换行运行：

```powershell
wslc run -d `
  --name bililive-helper-go `
  --restart always `
  -p 5689:5000 `
  -v D:\video\recordings:/recordings `
  -v D:\docker\bililive-helper-go:/data `
  -e TARGET_DIR=/recordings `
  -e LOG_DIR=/data `
  -e TZ=Asia/Shanghai `
  --cpus="3.0" `
  --memory="4g" `
  bililive-helper-go
```

> **Windows 路径支持**：
> - 宿主机卷挂载可直接使用 Windows 盘符路径（如 `D:\video\recordings:/recordings`）或 WSL 路径（如 `/mnt/d/video/recordings:/recordings`）。
> - 容器内部始终使用通用的 `/recordings` 和 `/data`，彻底解耦本地路径与容器配置。

---

### 3. 访问与初始化

启动成功后，浏览器访问：
```text
http://<宿主机IP>:5689
```
> 首次启动系统将自动进入 **初始化向导**，设置 Web 控制台密码后即可进入管理控制台。

## 环境依赖

| 依赖 | 版本 | 说明 |
|------|------|------|
| Go | 1.26+ | 后端编译（本地开发） |
| Node.js | 20.19+ | 前端编译（Vite 8 要求） |
| FFmpeg | 任意 | 视频转码、校验（Docker 镜像内已预装） |
| Docker | 20.10+ | 容器化部署（推荐） |

## 开发调试

```bash
# 后端（开发模式）
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
- [配置参考](docs/config.md) — 环境变量、目录挂载与 config.json 字段

## 常见问题

**ffmpeg 未安装** → Docker 镜像已内置 FFmpeg；若本地宿主机运行请通过包管理器安装。

**忘记密码** → 删除数据目录中的 `.credentials.json`（即 `{DATA_DIR}/.credentials.json`）后重启容器，系统将重新进入初始化向导。

**磁盘超阈值未触发清理** → 检查是否在静默时段（默认 04:00~12:00）、视频是否在安全期保护范围内，或该主播是否开启了「一键免删」。

**合并失败 "磁盘空间不足"** → TS 合并管线峰值需约 3 倍源文件大小空间，建议保证录制目录预留充足剩余容量。

## 许可证

[MIT License](LICENSE)
