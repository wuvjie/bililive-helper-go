# Bililive Helper

> B 站 / 抖音直播录制后处理工具 — 自动合并、智能清理、Web 管理控制台

[![Go](https://img.shields.io/badge/Go-1.26-blue.svg)](https://golang.org/)
[![Vue](https://img.shields.io/badge/Vue-3.5-green.svg)](https://vuejs.org/)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

## 项目简介

Bililive Helper 是一款面向 NAS 部署的直播录制后处理工具。它自动监控 [bililive-go](https://github.com/hr3lxpesr6/bililive-go) 的录制目录，将 FLV/MP4 分片合并为完整视频，在磁盘空间不足时智能清理旧文件，并提供一个基于 Notion 风格的 Web 控制台进行实时管理。

**技术栈**

| 层 | 技术 |
|----|------|
| 后端 | Go 1.26、Gin 1.12、go.uber.org/zap |
| 前端 | Vue 3.5、TypeScript 6、Element Plus 2.14、Pinia 3、Vue Router 5、Vite 8 |
| 视频处理 | FFmpeg（stream-copy 无损转码 + h264_rkmpp 硬件加速） |
| 部署 | Docker、Docker Compose（Alpine 多阶段构建） |

---

## 功能特性

### 录制合并
- FLV → MP4 无损转码（全程 stream-copy，零重编码）
- 多片段 TS 拼接合并（基于 ffprobe 探测实际时长，用前一片段结束时间计算间隔，精准划分场次）
- 硬件编码 fallback（h264_rkmpp / libx264，自动检测）
- **文件健康预检**：合并前通过 ffprobe 快速检测文件完整性，自动跳过损坏文件，避免浪费几十分钟的 ffmpeg 处理
- 原子写入 + 三重校验（ffprobe 元数据 + 多点解码测试 + 大小检查），失败时原始文件 100% 保留
- 文件写入检测（双快照对比），防止合并正在录制中的文件
- 磁盘空间预检查：可用空间不足时跳过合并，低于 10GB 硬限时跳过所有操作
- 每次合并创建独立操作日志文件（`op_{logID}.log`），同时输出到 SSE 实时流

### 智能清理
- 磁盘阈值触发（可配置触发 / 目标阈值）
- 白名单保护（关键词匹配，不分大小写）
- 安全期保护（按小时 / 按天两种模式）
- 已合并文件优先删除，每主播最少保留
- 单次删除上限控制，防止误删
- 正在写入的文件自动跳过（双快照检测）
- 支持 ctx 取消：SSE 断开或客户端断连时及时终止清理
- 每次清理创建独立操作日志文件（`op_{logID}.log`），同时输出到 SSE 实时流
- 磁盘空间硬限保护：可用空间低于 10GB 时跳过所有合并操作，防止磁盘写满
- **紧急清理**：可临时覆盖目标阈值快速释放空间（操作结束后自动恢复原配置）

### Web 控制台
- **系统概览**：磁盘状态、磁盘告警横幅、今日 / 本月统计、近 7 天趋势图、最近操作流水
- **主播管理**：主播列表、录制文件数、磁盘占用、最新视频时间
- **任务中心**：手动合并 / 清理、紧急清理、实时 SSE 输出终端（自动重连）
- **操作日志**：分页历史记录、按主播名搜索、per-operation 日志详情查看、导出 JSON
- **全局设置**：基本配置、存储管理、定时任务、系统诊断、智能推荐、配置导入导出（含文件选择器）

### 定时调度
- 可配置间隔的自动合并 / 清理
- **事件驱动调度**：目录变化检测（轮询主播目录修改时间），录制结束后自动触发合并，固定间隔作为兜底
- 静默时段（暂停任务窗口，支持跨午夜）
- **任务队列与并发控制**：可配置同时运行的最大任务数，避免多个 ffmpeg 进程同时打满磁盘 I/O
- 冷启动后立即触发首次任务
- 每日自动清理过期历史记录

### 安全设计
- bcrypt 密码哈希（常量时间比较，防时序攻击）
- API Token constant-time 比较（`crypto/subtle`）
- 登录频率限制（5 次 / 5 分钟 / IP）
- 全局 POST 请求频率限制（60 次 / 分钟 / IP，令牌桶算法）
- 可信代理限制（仅信任回环地址，防止 X-Forwarded-For 伪造绕过限流）
- Session HttpOnly + SameSite=Lax + 可选 Secure
- 改密后自动递增 `SessionVersion`，所有旧 Session 立即失效
- Content-Security-Policy 安全头（XSS 防护、点击劫持防护、CSP 策略）
- 路径穿越验证（`ValidateFilename` + `filepath.Clean` 前缀校验双重防护）
- 配置文件原子写入（`write → fsync → rename`，崩溃安全）
- 凭据文件独立存储（`0600` 权限，不写入 config.json）
- 历史记录损坏自动备份（`.corrupt.{timestamp}` 备份后重建空记录）

---

## 快速开始

### Docker Compose（推荐）

```bash
# 1. 克隆项目
git clone https://github.com/wuvjie/bililive-helper-go.git
cd bililive-helper-go

# 2. 修改 docker-compose.yml 中的录像目录（必须）
#    将 /bililive-go/recordings 改为你实际的 bililive-go 录像输出路径
#    例：/vol2/1000/video/bililive-go/抖音

# 3. 启动
docker compose up -d

# 4. 访问 Web 控制台（首次运行进入初始化向导）
open http://localhost:5689
```

> **注意：** 首次构建需要下载依赖，国内环境约 3-5 分钟，海外环境需在 docker-compose.yml 中添加 build-arg（见下方说明）。

首次启动后，访问 Web 控制台会自动进入初始化向导，引导设置登录密码。完成初始化后自动登录。

### 海外构建

默认使用国内镜像源（阿里云 + goproxy.cn + npmmirror）。海外环境需传入 build-arg 覆盖：

```bash
docker compose build \
  --build-arg ALPINE_MIRROR=dl-cdn.alpinelinux.org \
  --build-arg GO_PROXY=https://proxy.golang.org,direct \
  --build-arg NPM_REGISTRY=https://registry.npmjs.org
docker compose up -d
```

### 首次配置向导

启动后访问 Web 控制台，首次运行会自动进入初始化向导：

1. 设置登录密码（至少 6 位）
2. 完成初始化后自动登录

---

## 环境依赖

| 依赖 | 版本要求 | 说明 |
|------|----------|------|
| Go | 1.26+ | 后端编译 |
| Node.js | 20.19+ | 前端编译（Vite 8 要求） |
| FFmpeg / FFprobe | 任意稳定版 | 视频转码、校验 |
| Docker | 20.10+ | 容器化部署（推荐） |
| Docker Compose | 2.0+ | 编排部署 |

---

## 目录结构

```
bililive-helper-go/
├── cmd/server/
│   └── main.go                    # 入口（~15 行，委托给 app 包）
├── internal/
│   ├── app/
│   │   └── app.go                 # 应用生命周期（初始化、启动、优雅停机）
│   ├── config/
│   │   └── config.go              # 配置管理（原子写入、事务回滚、快照、DTO 自动化）
│   ├── fsutil/
│   │   ├── atomic.go              # 崩溃安全的文件操作（AtomicSave）
│   │   ├── scan.go                # 统一目录扫描（ScanStreamerDirs）
│   │   ├── safeio.go              # 安全删除（带重试）
│   │   └── path.go                # 路径安全校验
│   ├── handler/
│   │   ├── auth.go                # 登录、登出、密码修改、初始化向导
│   │   ├── config.go              # 配置读写、推荐、导入导出
│   │   ├── handler.go             # Handler 结构体定义与构造函数
│   │   ├── history.go             # 历史记录分页查询、操作日志查看
│   │   ├── merge.go               # 合并/清理任务执行
│   │   ├── response.go            # 统一响应辅助函数（ok/fail）
│   │   ├── routes.go              # 路由注册（集中管理）
│   │   ├── sse.go                 # SSE 流式传输辅助
│   │   └── status.go              # 系统状态、统计、主播列表
│   ├── middleware/
│   │   └── auth.go                # 认证中间件、频率限制、安全头
│   ├── model/
│   │   └── model.go               # 数据模型
│   ├── service/
│   │   ├── clean.go               # 智能清理服务
│   │   ├── dirwatcher.go          # 目录变化监听（事件驱动调度）
│   │   ├── history.go             # 历史记录持久化服务
│   │   ├── interfaces.go          # TaskRunner 接口定义
│   │   ├── merge.go               # 合并服务
│   │   ├── oplog.go               # 操作日志器
│   │   ├── scanner.go             # 文件扫描、分组、场次划分
│   │   ├── scheduler.go           # 定时调度器
│   │   ├── task_helper.go         # 任务启动通用逻辑（PrepareTask）
│   │   └── taskqueue.go           # 任务队列与并发控制
│   └── ffmpeg/
│       ├── ffmpeg.go              # FFmpeg 进程管理
│       ├── executor.go            # Executor 接口 + DefaultExecutor
│       ├── probe.go               # ffprobe 操作（ProbeDuration/ProbeHealth）
│       ├── concat.go              # TS 拼接合并
│       ├── convert.go             # FLV → MP4 转换
│       ├── reencode.go            # 重编码 fallback
│       └── validate.go            # 输出校验
├── frontend/                       # Vue 3 前端（Vite 8 构建）
│   └── src/
│       ├── api/                    # API 模块（8 个文件）
│       ├── composables/            # Vue composables（useFetchOnActivate 等）
│       ├── store/modules/          # Pinia stores（app、auth）
│       ├── views/                  # 页面组件（7 个视图）
│       ├── layout/                 # 布局组件（含子组件）
│       └── style/                  # 设计系统（token + 组件类）
├── docs/
│   ├── adr/                        # 架构决策记录
│   └── superpowers/specs/          # 设计文档
├── tests/baseline/                 # 重构安全网基线
├── templates/                      # 前端编译产物
├── Dockerfile                      # 多阶段构建
├── docker-compose.yml              # 部署配置
└── README.md
```

---

## API 接口说明

所有 `/api/*` 路由（除公开接口外）均需登录认证。认证方式支持 Session Cookie 和 Bearer Token。

### 公开接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/login` | 登录页面 |
| POST | `/api/login` | 用户登录（JSON: `{password}`） |
| GET | `/api/health` | 健康检查 |
| GET | `/api/setup/status` | 查询是否为首次运行 |
| POST | `/api/setup/init` | 首次初始化（设置密码） |

### 认证接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/auth/check` | 检查登录状态 |
| POST | `/api/auth/change-password` | 修改密码 |

### 系统状态

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/status` | 简要状态（磁盘使用率、主播数） |
| GET | `/api/status/detail` | 详细状态（磁盘、待合并文件、调度状态） |
| GET | `/api/stats` | 统计数据（今日 / 本月合并清理、7 天趋势） |
| GET | `/api/setup/check` | 系统诊断（FFmpeg、目录权限、磁盘空间） |

### 配置管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/config` | 获取当前配置（不含敏感字段） |
| POST | `/api/config` | 保存配置（部分更新） |
| GET | `/api/config/recommend` | 智能推荐配置（基于磁盘和内容分析） |
| GET | `/api/config/defaults` | 获取默认配置值 |
| GET | `/api/config/export` | 导出完整配置（含调度、历史） |
| POST | `/api/config/import` | 导入配置 |

### 定时调度

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/schedule` | 获取调度状态（间隔、启用状态、上次 / 下次执行时间） |
| POST | `/api/schedule` | 保存调度配置 |
| POST | `/api/schedule/run/:task` | 手动触发任务（`merge` 或 `clean`） |

### 主播与文件

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/streamers` | 获取主播列表（文件数、磁盘占用、最新视频时间） |
| GET | `/api/streamers/:name/files` | 获取指定主播的文件列表 |

### 任务执行（SSE 流式输出）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/merge` | 执行合并（可指定主播，SSE 流式返回进度） |
| POST | `/api/merge/manual` | 手动合并指定文件 |
| POST | `/api/merge/retry` | 重试合并 |
| POST | `/api/clean` | 执行清理（可指定主播，SSE 流式返回进度） |
| GET | `/api/clean/estimate` | 预估可清理文件数和大小 |
| POST | `/api/clean/emergency` | 紧急清理（可自定义目标阈值） |

### 历史与日志

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/history` | 分页查询历史记录（支持 `task`、`streamer` 过滤） |
| GET | `/api/history/export` | 导出全部历史记录 |
| GET | `/api/logs/content/:task` | 获取操作日志内容（查询参数 `log_id`，返回最近 200 行） |

---

## 配置说明

### 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `PASSWORD` | 登录密码（可选，优先通过初始化向导设置；适用于自动化部署） | 首次启动自动生成 |
| `SECRET_KEY` | Session 签名密钥（可选，通常无需设置） | 首次启动自动生成 |
| `TARGET_DIR` | 录制视频目录 | `/vol2/1000/video/bililive-go/抖音` |
| `LOG_DIR` | 日志和配置存储目录 | `/vol1/1000/docker/bililive-helper-go` |
| `TZ` | 时区 | `Asia/Shanghai` |
| `COOKIE_SECURE` | Session cookie Secure 标志（HTTPS 时设为 true） | `false` |
| `WEBHOOK_URL` | 任务完成通知 webhook（JSON POST） | 空（不通知） |
| `API_TOKEN` | API Token（用于 Bearer 认证） | 空（未启用） |
| `PORT` | 监听端口（也可通过 config.json 设置） | `5000` |

### 配置文件（config.json）

配置文件位于 `{LOG_DIR}/config.json`，首次初始化时自动生成。环境变量优先于配置文件。注意：`PASSWORD` 和 `SECRET_KEY` 使用 `json:"-"` 标记，不写入 config.json；密码单独持久化到 `{LOG_DIR}/.credentials.json`（`0600` 权限）。配置文件通过原子写入（`write → fsync → rename`）保证崩溃安全。

| 字段 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| `TARGET_DIR` | string | 录制视频目录 | `/vol2/1000/video/bililive-go/抖音` |
| `TRIGGER_THRESHOLD` | float | 清理触发阈值（磁盘使用率 %） | 80 |
| `TARGET_THRESHOLD` | float | 清理目标阈值（磁盘使用率 %） | 60 |
| `MIN_KEEP_PER_STREAMER` | int | 每主播最少保留文件数 | 3 |
| `SAFE_AGE_MINUTES` | int | 安全期（小时模式下的分钟数） | 120 |
| `GAP_MINUTES` | int | 场次间隔（分钟，超过此间隔视为新场次，基于 ffprobe 结束时间计算） | 30 |
| `MERGE_AGE_MINUTES` | int | 合并等待时间（文件落盘后等待分钟数） | 30 |
| `WHITELIST_KEYWORDS` | []string | 白名单关键词（匹配的文件不会被清理） | `["留存","纪念","高能","生日","勿删"]` |
| `SAFE_MODE` | string | 安全期模式：`hours` 或 `days` | `hours` |
| `SAFE_DAYS` | int | 安全期天数（仅 days 模式生效） | 1 |
| `MAX_DELETE_PER_RUN` | int | 单次清理最大删除数 | 10 |
| `BACKUP_START_HOUR` | int | 静默时段开始小时 | 4 |
| `BACKUP_START_MINUTE` | int | 静默时段开始分钟 | 0 |
| `BACKUP_END_HOUR` | int | 静默时段结束小时 | 12 |
| `BACKUP_END_MINUTE` | int | 静默时段结束分钟 | 0 |
| `PORT` | int | 监听端口 | 5000 |
| `SESSION_VERSION` | int | Session 版本号（改密时自动递增，使旧 Session 失效） | 0 |

### 资源限制（NAS 部署参考）

```yaml
deploy:
  resources:
    limits:
      cpus: '3.0'    # 4 核留 1 核给录制进程
      memory: 4G     # 8G 留 4G 给录制进程
```

---

## 开发指南

### 本地开发

```bash
# 后端（默认监听 :5000）
go run ./cmd/server

# 前端（开发模式，热更新，API 代理到 :5000）
cd frontend
npm install
npm run dev
# 访问 http://localhost:3000
```

### 构建部署

```bash
# 编译前端
cd frontend && npm run build

# 编译后端
cd .. && go build -o bililive-helper-go ./cmd/server

# Docker 一键构建
docker compose up -d --build
```

### 运行测试

```bash
# 后端测试
go test ./... -v -count=1

# 前端测试
cd frontend && npm test

# 前端 lint
cd frontend && npm run lint
```

### 日志格式

每次手动或自动执行的合并/清理操作都会创建独立的操作日志文件 `op_{logID}.log`，存储在 `{LOG_DIR}/{task}_log/` 目录下（如 `merge_log/`、`clean_log/`）。日志 ID 格式为 `{task}_{YYYYMMDD}_{HHMMSS}_{4位hex}`，可通过操作日志页面的「查看」按钮查看详细输出。超过 30 天的操作日志文件在清理任务中自动删除。

```
▶ 开始 [全局] 合并
⚙ 扫描 /path ...
── 主播A ──
[主播A] ⏭ 3个文件 → 主播正在录制，跳过
── 主播B ──
[主播B] ⚙ 合并 8 个文件 (0.8 GB)
🔄 转换 001.flv → TS…
🔄 转换 002.flv → TS…
⚙ 拼接 TS 文件…
[主播B] ✅ → 合并版.mp4 (8 个文件)
───────────────────────────
✅ 完成: 扫描 12 个主播, 合并 1 场次 (0.8 GB)
```

---

## 常见问题

### Q: 启动时提示 "ffmpeg 未安装或不在 PATH 中"

安装 FFmpeg 后确保其在系统 PATH 中：
```bash
# Ubuntu/Debian
apt install ffmpeg

# Alpine（Docker 环境）
apk add ffmpeg

# 验证
ffmpeg -version
```

### Q: 首次启动密码忘记了

删除 `{LOG_DIR}/.credentials.json` 文件后重启，会自动生成新密码并打印到控制台。

### Q: 磁盘使用率已超过阈值但没有触发清理

检查以下几点：
1. 确认 `TRIGGER_THRESHOLD` 设置合理（默认 80%）
2. 确认自动清理已启用（调度设置中的 `clean_enabled`）
3. 确认当前不在静默时段（默认 04:00-12:00）
4. 检查文件是否在安全期内（默认 120 分钟内不清理）
5. 检查文件是否被白名单关键词保护

### Q: 合并任务一直显示 "落盘等待"

这是正常行为。系统会等待文件停止写入并且经过 `MERGE_AGE_MINUTES`（默认 30 分钟）后才开始合并，以确保录制完全结束。

### Q: 合并失败，显示 "磁盘空间不足"

TS 合并管线在峰值时需要约 3 倍源文件大小的磁盘空间（源文件 + TS 中间文件 + 输出文件 + 2GB 余量）。请确保有足够可用空间，或减小单次合并的文件大小。

### Q: 如何排除特定主播的文件不被清理

在配置的白名单关键词（`WHITELIST_KEYWORDS`）中添加主播名称或文件名中的关键词即可。匹配逻辑不区分大小写。

### Q: Webhook 通知不生效

确认 `WEBHOOK_URL` 环境变量已设置且指向一个可访问的 JSON POST 端点。请求体格式为 `{"text": "[Bililive Helper] 消息内容"}`。

### Q: 修改密码后，其他设备上的登录会怎样

修改密码后，系统自动递增 `SessionVersion`，所有旧的登录会话（包括其他浏览器或设备上的）会立即失效，需要重新输入新密码登录。无需手动清理 cookie。

### Q: Docker 容器内 FFmpeg 硬件加速不生效

当前仅支持 Rockchip 平台的 `h264_rkmpp` 硬件编码器，系统会自动检测并降级到 `libx264`。如需其他硬件加速支持，需修改 `internal/ffmpeg/reencode.go` 中的编码器检测逻辑。

---

## 许可证

[MIT License](LICENSE)
