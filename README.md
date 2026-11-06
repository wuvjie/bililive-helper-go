# Bililive Helper Go

> B 站 / 抖音直播录制后处理工具 — 自动合并、智能清理、Web 管理控制台

[![Version](https://img.shields.io/badge/Version-v3.0.0-blue.svg)](https://github.com/wuvjie/bililive-helper-go)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## 简介

自动监控 [bililive-go](https://github.com/hr3lxpesr6/bililive-go) 等录制工具的输出目录，自动将录制生成的 FLV/TS 视频分片无损拼接或转换为完整 MP4 视频；当磁盘空间不足时，自动根据保留策略清理过期录像；同时提供功能完善的 Web 管理控制台，支持在线管理主播、视频回放、配置调整及实时日志查看。

## 核心功能

- **录制合并与转码**：
  - FLV 转 MP4 无损快速转封装，TS 分片自动流拷贝拼接；
  - 录制时间戳精确保持，关联的 XML 弹幕及封面图片自动重命名对齐。
- **自动化视频裁剪**：
  - 支持按主播配置时段裁剪规则（如自动截取开播前 30 分钟），转码时自动裁剪并替换，免去人工二次剪辑。
- **智能磁盘清理**：
  - 磁盘空间达到警戒线时自动按需清理旧视频，支持免删保护期（按小时或天数保护最新录像）；
  - 支持主播级独立策略（可单独为某个主播设置保留天数、最少保留视频数、或一键免删保护）；
  - 清理视频时级联删除对应的 XML 弹幕及封面文件，不留孤岛文件。
- **自动化任务与监控**：
  - 录制目录变动实时感知（录制完成自动即刻触发）+ 定时巡检双重保障；
  - 任务并发互斥控制，支持设置避让时段（如每日数据备份期暂停任务）。
- **Web 管理控制台**：
  - **系统概览**：磁盘空间实时监控、近期产出与清理趋势图表、系统健康状态；
  - **主播管理**：主播列表/卡片双视图、录像文件抽屉管理、在线视频播放预览；
  - **任务中心**：实时任务控制台日志输出、历史任务归档分类回查（支持按类型筛选）、日志关键词过滤；
  - **配置与备份**：所有全局策略与主播个性化规则支持一键导出为 JSON 备份及导入恢复；
  - **使用体验**：支持日常模式与暗黑夜间模式切换，支持常用键盘快捷键（`/` 快速聚焦搜索、`D` 切换暗黑模式、`Esc` 关闭弹层等）。
- **账户与安全**：
  - 管理员密码加密存储、连续登录失败冷却保护、会话安全隔离、路径穿越校验。

---

## 安装部署

本工具推荐通过容器化方式部署（开箱即用，内置 FFmpeg 与所有运行时），支持 **Docker Compose**、**原生 Docker CLI (`docker run`)** 以及 **Windows WSL Container (`wslc`)** 三种方式。

无论哪种部署方式，容器内部均采用统一的标准固定路径（实现宿主机与容器的彻底解耦）：
- `/recordings`：录制视频根目录（宿主机挂载 bililive-go 等工具的录制输出目录）
- `/data`：持久化数据根目录（宿主机挂载配置、日志、会话凭据目录）

首先克隆项目：
```bash
git clone https://github.com/wuvjie/bililive-helper-go.git
cd bililive-helper-go
```

请选择最适合您运行环境的方式：

---

### 方式一：Docker Compose 部署（推荐，适合常规多容器与 NAS 环境）

适合习惯通过 YAML 文件和环境配置文件集中管理服务的用户。

#### 1. 配置宿主机真实路径
复制配置模板并编辑：
```bash
cp .env.example .env
```
在 `.env` 中填入您宿主机的实际路径（`.env` 已加入 `.gitignore`，保护您的宿主机隐私信息不被提交）：
```env
# 录制视频根目录（宿主机实际路径）
RECORDINGS_DIR=/path/to/recordings

# 配置文件与日志数据持久化目录（宿主机实际路径）
DATA_DIR=/path/to/data

# Web 访问端口（默认 5689）
PORT=5689
```

#### 2. 构建镜像
- **标准构建（默认使用官方标准源）**：
  ```bash
  docker compose build
  ```
- **国内网络加速构建（若直连下载较慢）**：
  ```bash
  docker compose build \
    --build-arg REGISTRY=docker.m.daocloud.io \
    --build-arg ALPINE_MIRROR=mirrors.aliyun.com \
    --build-arg GO_PROXY=https://goproxy.cn,direct \
    --build-arg NPM_REGISTRY=https://registry.npmmirror.com
  ```

#### 3. 启动与管理
```bash
docker compose up -d          # 后台启动容器
docker compose logs -f        # 查看实时运行日志
docker compose restart        # 重启服务
docker compose down           # 停止并移除容器
```

---

### 方式二：Docker CLI 原生命令行（适合 Linux / NAS / macOS 无 Compose 环境）

适合偏好轻量单条命令运行、不希望安装或维护 docker-compose 的用户。

#### 1. 构建镜像 (`docker build`)
在项目根目录下执行：
- **标准构建（默认官方源）**：
  ```bash
  docker build -t bililive-helper-go .
  ```
- **国内网络加速构建**：
  ```bash
  docker build -t bililive-helper-go \
    --build-arg REGISTRY=docker.m.daocloud.io \
    --build-arg ALPINE_MIRROR=mirrors.aliyun.com \
    --build-arg GO_PROXY=https://goproxy.cn,direct \
    --build-arg NPM_REGISTRY=https://registry.npmmirror.com \
    .
  ```

#### 2. 运行容器 (`docker run`)
通过 `-v` 挂载将宿主机真实目录对应到容器内部标准路径：
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
  bililive-helper-go
```

> **参数说明**：
> - `-v /path/to/recordings:/recordings`：将宿主机录像根目录（如 `/path/to/recordings` 或 NAS 的 `/volume1/...`）映射到容器 `/recordings`。
> - `-v /path/to/data:/data`：将宿主机数据持久化目录映射到容器 `/data`。
> - `-p 5689:5000`：宿主机对外暴露 5689 端口（容器内监听 5000）。
> - `-e TARGET_DIR=/recordings -e LOG_DIR=/data`：确保应用程序使用挂载的标准目录。
> - *可选参数*：若需限制硬件占用或日志体积，可按需追加 `--cpus="2.0" --memory="4g"` 或 `--log-opt max-size=10m --log-opt max-file=3`。

#### 3. 运维维护
```bash
docker logs -f bililive-helper-go     # 查看实时日志
docker restart bililive-helper-go    # 重启容器
docker stop bililive-helper-go       # 停止容器
```

---

### 方式三：WSL Container 部署（适合 Windows PowerShell + `wslc`）

适合在 Windows 环境下使用 WSL Container CLI (`wslc`) 进行本地或测试部署的用户。原生支持 PowerShell 换行语法（反引号 `）与 Windows 盘符路径。

#### 1. 构建镜像 (`wslc build`)
在 PowerShell 中进入项目目录执行：
- **标准构建（默认官方源）**：
  ```powershell
  wslc build -t bililive-helper-go .
  ```
- **国内网络加速构建**：
  ```powershell
  wslc build -t bililive-helper-go `
    --build-arg REGISTRY=docker.m.daocloud.io `
    --build-arg ALPINE_MIRROR=mirrors.aliyun.com `
    --build-arg GO_PROXY=https://goproxy.cn,direct `
    --build-arg NPM_REGISTRY=https://registry.npmmirror.com `
    .
  ```

#### 2. 运行容器 (`wslc run`)
```powershell
wslc run -d `
  --name bililive-helper-go `
  -p 5689:5000 `
  -v D:\video\recordings:/recordings `
  -v D:\docker\bililive-helper-go:/data `
  -e TARGET_DIR=/recordings `
  -e LOG_DIR=/data `
  -e TZ=Asia/Shanghai `
  bililive-helper-go
```

> **说明与参数特性**：
> - **路径映射**：宿主机路径既支持 Windows 原生盘符路径（如 `D:\video\recordings:/recordings`），也支持 WSL 挂载路径（如 `/mnt/d/video/recordings:/recordings`）。容器内部始终固定使用 `/recordings` 和 `/data`。
> - **重启策略**：`wslc` 暂不支持 `--restart` 选项；容器停止后可随时使用 `wslc start bililive-helper-go` 快速拉起。
> - **资源限制（可选）**：若需限制硬件占用，可追加 `--cpus="2.0" -m="4g"`。

#### 3. 运维维护
```powershell
wslc logs -f bililive-helper-go     # 查看实时日志
wslc restart bililive-helper-go    # 重启容器
wslc stop bililive-helper-go       # 停止容器
wslc start bililive-helper-go      # 启动已停止的容器
```

---

## Web 访问与初始化向导

无论采用上述哪种方式启动，服务就绪后直接在浏览器中打开：
```text
http://<宿主机IP>:5689
```
- **首次启动**：系统将自动进入 **初始化向导**，设置 Web 控制台管理员密码后即可进入管理控制台。
- **凭据重置**：若遗忘密码，删除宿主机数据目录下 `{DATA_DIR}/.credentials.json` 后重启容器，系统将重新进入初始化向导。

---

## 常见问题与说明

**Q: 录像合并需要安装什么工具吗？**  
A: 不需要。Docker 镜像中已经内置了完整可靠的 FFmpeg 视频处理流水线，开箱即用。

**Q: 忘记管理员密码怎么办？**  
A: 删除宿主机挂载数据目录中的 `.credentials.json`（即 `{DATA_DIR}/.credentials.json`），然后重启容器，系统会自动重新进入首次启动向导让您设定新密码。

**Q: 磁盘达到清理阈值了，为什么有些录像没有被清理？**  
A: 请排查以下原因：
1. 视频处于免删保护期内（默认对最新录像提供免删保护，可在全局策略中调整）；
2. 该主播开启了「一键免删保护」或设置了较高的「最少保留视频数」；
3. 当前处于设定的静默时段（如备份避让期）。

**Q: 提示合并失败 "磁盘空间不足"？**  
A: 视频分片流拷贝拼接过程中需要临时工作空间，建议录制目录预留充裕的剩余可用空间。

## 许可证

[MIT License](LICENSE)
