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
| 部署 | Docker Compose、Docker CLI (`docker run`)、WSL Container (`wslc`) |

## 功能

- **录制合并**：FLV→MP4 无损转码、TS Concat Demuxer 稳定拼接、时间戳对齐、XML 弹幕与封面伴生对齐
- **智能清理**：磁盘阈值触发、白名单保护、安全期保护、一键免删、伴生弹幕与封面级联清理
- **任务调度**：事件驱动（目录变化检测）+ 固定间隔兜底、任务队列并发控制、静默时段
- **Web 控制台**：系统概览、主播管理、文件抽屉与在线视频预览、任务中心（SSE 实时日志）、操作审计、全局策略与备份
- **安全设计**：bcrypt 哈希、登录限流、Session 版本化、CSP 安全头、路径穿越防护

---

## 部署运行

本工具支持 **Docker Compose**、**原生 Docker (`docker run`)** 以及 **Windows WSL Container (`wslc`)** 三种部署方式。

无论哪种部署方式，容器内部均采用统一的标准固定路径（实现宿主机与容器的彻底解耦）：
- `/recordings`：录制视频根目录（宿主机挂载 bililive-go 录制输出目录）
- `/data`：持久化数据根目录（宿主机挂载配置、日志、会话凭据目录）

首先克隆项目：
```bash
git clone https://github.com/wuvjie/bililive-helper-go.git
cd bililive-helper-go
```

请选择最适合您运行环境的方式：

---

### 方案一：Docker Compose 部署（推荐，适合常规多容器与 NAS 环境）

适合习惯通过 YAML 文件和环境配置文件集中管理服务的用户。

#### 1. 配置宿主机真实路径
复制配置模板并编辑：
```bash
cp .env.example .env
```
在 `.env` 中填入您宿主机的实际路径（`.env` 已加入 `.gitignore`，保护您的宿主机隐私信息不被提交）：
```env
# 录制视频根目录（宿主机实际路径）
RECORDINGS_DIR=/volume1/video/recordings

# 配置文件与日志数据持久化目录（宿主机实际路径）
DATA_DIR=/volume1/docker/bililive-helper-go

# Web 访问端口（默认 5689）
PORT=5689
```

#### 2. 构建镜像
- **国内网络（默认加速源）**：
  ```bash
  docker compose build
  ```
- **海外 VPS / 直连官方源**：
  ```bash
  docker compose build \
    --build-arg REGISTRY=docker.io \
    --build-arg ALPINE_MIRROR=dl-cdn.alpinelinux.org \
    --build-arg GO_PROXY=https://proxy.golang.org,direct \
    --build-arg NPM_REGISTRY=https://registry.npmjs.org
  ```

#### 3. 启动与管理
```bash
docker compose up -d          # 后台启动容器
docker compose logs -f        # 查看实时运行日志
docker compose restart        # 重启服务
docker compose down           # 停止并移除容器
```

---

### 方案二：Docker CLI 原生命令行（适合 Linux / NAS / macOS 无 Compose 环境）

适合偏好轻量单条命令运行、不希望安装或维护 docker-compose 的用户。

#### 1. 构建镜像 (`docker build`)
在项目根目录下执行：
- **国内网络（默认加速源）**：
  ```bash
  docker build -t bililive-helper-go .
  ```
- **海外 VPS / 直连官方源**：
  ```bash
  docker build -t bililive-helper-go \
    --build-arg REGISTRY=docker.io \
    --build-arg ALPINE_MIRROR=dl-cdn.alpinelinux.org \
    --build-arg GO_PROXY=https://proxy.golang.org,direct \
    --build-arg NPM_REGISTRY=https://registry.npmjs.org .
  ```

#### 2. 运行容器 (`docker run`)
通过 `-v` 挂载将宿主机真实目录对应到容器内部标准路径：
```bash
docker run -d \
  --name bililive-helper-go \
  --restart always \
  -p 5689:5000 \
  -v /volume1/video/recordings:/recordings \
  -v /volume1/docker/bililive-helper-go:/data \
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
> - `-v /volume1/video/recordings:/recordings`：将宿主机 bililive-go 录制根目录映射到容器 `/recordings`。
> - `-v /volume1/docker/bililive-helper-go:/data`：将宿主机配置、凭据与日志持久化目录映射到容器 `/data`。
> - `-p 5689:5000`：宿主机对外暴露 5689 端口（容器内监听 5000）。
> - `-e TARGET_DIR=/recordings -e LOG_DIR=/data`：确保应用程序使用挂载的标准目录。

#### 3. 运维维护
```bash
docker logs -f bililive-helper-go     # 查看实时日志
docker restart bililive-helper-go    # 重启容器
docker stop bililive-helper-go       # 停止容器
```

---

### 方案三：WSL Container 部署（适合 Windows PowerShell + `wslc`）

适合在 Windows 环境下使用 WSL Container CLI (`wslc`) 进行本地或测试部署的用户。原生支持 PowerShell 换行语法（反引号 `）与 Windows 盘符路径。

#### 1. 构建镜像 (`wslc build`)
在 PowerShell 中进入项目目录执行：
- **国内网络（默认加速源）**：
  ```powershell
  wslc build -t bililive-helper-go .
  ```
- **海外 / 直连官方源**：
  ```powershell
  wslc build -t bililive-helper-go `
    --build-arg REGISTRY=docker.io `
    --build-arg ALPINE_MIRROR=dl-cdn.alpinelinux.org `
    --build-arg GO_PROXY=https://proxy.golang.org,direct `
    --build-arg NPM_REGISTRY=https://registry.npmjs.org .
  ```

#### 2. 运行容器 (`wslc run`)
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

> **Windows 路径挂载特别说明**：
> - 宿主机路径既支持 Windows 原生盘符路径（如 `D:\video\recordings:/recordings`），也支持 WSL 挂载路径（如 `/mnt/d/video/recordings:/recordings`）。
> - 容器内部始终使用通用的 `/recordings` 和 `/data`，彻底解耦本地路径与容器业务逻辑。

#### 3. 运维维护
```powershell
wslc logs -f bililive-helper-go     # 查看实时日志
wslc restart bililive-helper-go    # 重启容器
wslc stop bililive-helper-go       # 停止容器
```

---

### Web 访问与初始化向导

无论采用上述哪种方式启动，服务就绪后直接在浏览器中打开：
```text
http://<宿主机IP>:5689
```
- **首次启动**：系统将自动进入 **初始化向导**，设置 Web 控制台管理员密码后即可进入管理控制台。
- **凭据重置**：若遗忘密码，删除宿主机数据目录下 `{DATA_DIR}/.credentials.json` 后重启容器，系统将重新进入初始化向导。

---

## 环境依赖

| 依赖 | 版本 | 说明 |
|------|------|------|
| Go | 1.26+ | 后端编译（本地开发） |
| Node.js | 20.19+ | 前端编译（Vite 8 要求） |
| FFmpeg | 任意 | 视频转码、校验（Docker 镜像内已预装） |
| Docker / WSL | 20.10+ | 容器化部署（推荐） |

## 开发调试

```bash
# 后端（本地开发模式）
go run ./cmd/server

# 前端（本地开发模式，API 代理到 :5000）
cd frontend && npm install && npm run dev

# 测试
go test ./...           # 后端单元测试
cd frontend && npm test # 前端单元测试
cd frontend && npm run lint # 代码检查
```

## 文档

- [API 接口文档](docs/api.md) — 所有 REST 端点说明
- [配置参考](docs/config.md) — 目录挂载规范、环境变量与 config.json 字段

## 常见问题

**ffmpeg 未安装** → Docker 镜像已内置 FFmpeg；若本地宿主机运行请通过包管理器安装。

**忘记密码** → 删除宿主机数据目录中的 `.credentials.json`（即 `{DATA_DIR}/.credentials.json`）后重启容器，系统将重新进入初始化向导。

**磁盘超阈值未触发清理** → 检查是否在静默时段（默认 04:00~12:00）、视频是否在安全期保护范围内，或该主播是否开启了「一键免删」。

**合并失败 "磁盘空间不足"** → TS 合并管线峰值需约 3 倍源文件大小空间，建议保证录制目录预留充足剩余容量。

## 许可证

[MIT License](LICENSE)
