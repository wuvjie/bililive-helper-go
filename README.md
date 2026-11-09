# Bililive Helper Go

> 通用直播录像后处理工具：视频分片合并、磁盘清理、Web 管理控制台。

[![Version](https://img.shields.io/badge/Version-v2.0.1-blue.svg)](CHANGELOG.md)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Docker Image](https://img.shields.io/badge/Docker%20Hub-wuvjie%2Fbililive--helper--go-blue?logo=docker)](https://hub.docker.com/r/wuvjie/bililive-helper-go)
[![Architecture](https://img.shields.io/badge/arch-amd64%20%7C%20arm64-brightgreen.svg)](https://hub.docker.com/r/wuvjie/bililive-helper-go)
[![Build Status](https://img.shields.io/github/actions/workflow/status/wuvjie/bililive-helper-go/docker-publish.yml?branch=main&label=CI%2FCD)](https://github.com/wuvjie/bililive-helper-go/actions/workflows/docker-publish.yml)

<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="docs/images/dashboard-dark.png">
    <source media="(prefers-color-scheme: light)" srcset="docs/images/dashboard-light.png">
    <img alt="Bililive Helper Go 控制台概览" src="docs/images/dashboard-light.png" width="100%">
  </picture>
</p>

---

## 功能

- **分片拼接**：监控录像目录，将录制生成的 FLV/TS 分片无损拼接为 MP4，自动对齐重命名关联的 XML 弹幕与封面文件并保留时间戳。
- **磁盘清理**：磁盘使用率达到设定阈值时按天数清理旧录像，级联删除弹幕与封面，支持按主播设置免删保护。
- **主播管理**：支持按主播设置录像时段裁剪与保留规则，支持在网页端直接预览播放录像。
- **控制台**：提供 Web 界面查看磁盘容量、运行实时任务终端、回查操作日志及切换深色模式。
- **环境集成**：镜像内置 FFmpeg，支持 Docker Compose、原生 Docker CLI 与 Windows WSL。

---

## 目录说明

容器内固定使用以下两个标准目录：

| 容器内路径 | 说明 | 宿主机挂载内容 |
| :--- | :--- | :--- |
| `/recordings` | 录像存储目录 | 挂载 bililive-go 等录制工具的输出文件夹 |
| `/data` | 数据存储目录 | 挂载配置文件（config.json）、运行日志与凭据 |

---

## 安装与部署

官方 Docker 镜像已发布至 Docker Hub（`wuvjie/bililive-helper-go`）及 GitHub Container Registry（`ghcr.io/wuvjie/bililive-helper-go`），内置 FFmpeg，支持 `linux/amd64` 与 `linux/arm64` 双架构。

### 方式一：Docker Compose（推荐）

新建一个 `docker-compose.yml` 文件（**若已克隆本仓库，可直接使用根目录下现成的 `docker-compose.yml`**）：

```yaml
services:
  bililive-helper-go:
    image: wuvjie/bililive-helper-go:latest
    container_name: bililive-helper-go
    restart: always
    ports:
      - "5689:5000"
    volumes:
      - /path/to/recordings:/recordings # 录像存储目录（如 bililive-go 输出路径）
      - /path/to/data:/data             # 配置文件与数据持久化目录
    environment:
      - TZ=Asia/Shanghai
```

在配置文件所在目录下启动容器：
```bash
docker compose up -d
```

---

### 方式二：Docker CLI

无需下载源码，直接拉取并运行官方镜像：

```bash
docker run -d \
  --name bililive-helper-go \
  --restart always \
  -p 5689:5000 \
  -v /path/to/recordings:/recordings \
  -v /path/to/data:/data \
  -e TZ=Asia/Shanghai \
  wuvjie/bililive-helper-go:latest
```

---

### 方式三：Windows WSL / WSL Container (`wslc`)

在 Windows PowerShell 环境下运行：

```powershell
wslc run -d `
  --name bililive-helper-go `
  -p 5689:5000 `
  -v "D:\video\recordings:/recordings" `
  -v "D:\docker\bililive-helper-go:/data" `
  wuvjie/bililive-helper-go:latest
```

---

### 开发者模式：源码本地构建

如需对前端或后端代码进行二次开发，可克隆代码仓并在本地编译镜像：

```bash
git clone https://github.com/wuvjie/bililive-helper-go.git
cd bililive-helper-go
cp .env.example .env
# 编辑 .env 配置后构建并启动
docker compose up -d --build
```

---

## 使用说明

服务启动后，通过浏览器访问：
```text
http://<宿主机IP>:5689
```

- **初始化向导**：系统无硬编码初始密码。首次启动访问网页时会自动进入初始化向导，由管理员在网页上自主设置登录密码（至少 6 位）。
- **重置密码**：若遗忘密码，删除宿主机挂载数据目录中的 `{DATA_DIR}/.credentials.json` 文件并重启容器，系统将重新进入初始化设置向导。
- **关于 .env 文件**：普通用户可直接在 `docker-compose.yml` 或 `docker run` 命令中填写实际挂载路径，无需额外创建 `.env` 文件；亦可按需创建 `.env` 文件进行路径变量抽离。

---

## 文档

- [更新日志 (CHANGELOG.md)](CHANGELOG.md) — 版本发布与演进记录
- [API 文档 (docs/api.md)](docs/api.md) — REST API 接口定义
- [配置说明 (docs/config.md)](docs/config.md) — 目录与参数配置详解

---

## 致谢

本项目代码主要由 AI 生成与构建，感谢以下大语言模型提供的支持：

> **Gemini** · **MiMo** · **DeepSeek** · **Claude** · **ChatGPT** · **Doubao**

---

## 许可证

本项目基于 [MIT License](LICENSE) 开源。
