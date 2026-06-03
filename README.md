# Bililive Helper

> B 站 / 抖音直播录制后处理工具 — 自动合并、智能清理、Web 管理控制台

[![Go](https://img.shields.io/badge/Go-1.25-blue.svg)](https://golang.org/)
[![Vue](https://img.shields.io/badge/Vue-3.5-green.svg)](https://vuejs.org/)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

## 项目简介

Bililive Helper 是一款面向 NAS 部署的直播录制后处理工具。它自动监控 [bililive-go](https://github.com/hr3lxpesr6/bililive-go) 的录制目录，将 FLV/MP4 分片合并为完整视频，在磁盘空间不足时智能清理旧文件，并提供一个基于 Notion 风格的 Web 控制台进行实时管理。

**技术栈**

| 层 | 技术 |
|----|------|
| 后端 | Go 1.25、Gin、go.uber.org/zap |
| 前端 | Vue 3.5、TypeScript、Element Plus、Pinia、Vite 6 |
| 视频处理 | FFmpeg（stream-copy 无损转码 + h264_rkmpp 硬件加速） |
| 部署 | Docker、Docker Compose（Alpine 多阶段构建） |

---

## 功能特性

### 录制合并
- FLV → MP4 无损转码（全程 stream-copy，零重编码）
- 多片段 TS 拼接合并（基于 ffprobe 探测实际时长，用前一片段结束时间计算间隔，精准划分场次）
- 硬件编码 fallback（h264_rkmpp / libx264，自动检测）
- 原子写入 + 三重校验（ffprobe 元数据 + 多点解码测试 + 大小检查），失败时原始文件 100% 保留
- 文件写入检测（双快照对比），防止合并正在录制中的文件

### 智能清理
- 磁盘阈值触发（可配置触发 / 目标阈值）
- 白名单保护（关键词匹配，不分大小写）
- 安全期保护（按小时 / 按天两种模式）
- 已合并文件优先删除，每主播最少保留
- 单次删除上限控制，防止误删
- 正在写入的文件自动跳过（双快照检测）

### Web 控制台
- **系统概览**：磁盘状态、今日 / 本月统计、近 7 天趋势图、最近操作流水
- **主播管理**：主播列表、录制文件数、磁盘占用
- **任务中心**：手动合并 / 清理、实时 SSE 输出终端
- **操作日志**：分页历史记录、日志详情查看、导出 JSON
- **全局设置**：基本配置、存储管理、定时任务、系统诊断、智能推荐、配置导入导出

### 定时调度
- 可配置间隔的自动合并 / 清理
- 静默时段（暂停任务窗口，支持跨午夜）
- 冷启动后立即触发首次任务
- 每日自动清理过期历史记录

### 安全设计
- bcrypt 密码哈希（运行时比较）
- API Token constant-time 比较（防时序攻击）
- 登录频率限制（5 次 / 5 分钟）
- 全局 POST 请求频率限制（60 次 / 分钟 / IP）
- Session HttpOnly + SameSite=Lax + 可选 Secure
- Content-Security-Policy 安全头
- 路径穿越验证、输入参数校验

---

## 快速开始

### Docker Compose（推荐）

```bash
# 克隆项目
git clone <repo-url> bililive-helper-go
cd bililive-helper-go

# 启动
docker compose up -d

# 访问 Web 控制台（首次运行进入初始化向导）
open http://localhost:5689
```

首次启动后，访问 Web 控制台会自动进入初始化向导，引导设置登录密码。录制目录和日志目录通过 docker-compose.yml 的环境变量配置。完成初始化后自动登录。

### 首次配置向导

启动后访问 Web 控制台，首次运行会自动进入初始化向导：

1. 设置登录密码（至少 6 位）
2. 完成初始化后自动登录

---

## 环境依赖

| 依赖 | 版本要求 | 说明 |
|------|----------|------|
| Go | 1.25+ | 后端编译 |
| Node.js | 18+ | 前端编译 |
| FFmpeg / FFprobe | 任意稳定版 | 视频转码、校验 |
| Docker | 20.10+ | 容器化部署（推荐） |
| Docker Compose | 2.0+ | 编排部署 |

---

## 目录结构

```
bililive-helper-go/
├── cmd/server/
│   └── main.go                    # 入口：路由注册、中间件、优雅停机
├── internal/
│   ├── config/
│   │   └── config.go              # 配置管理（原子写入、事务回滚、快照）
│   ├── handler/
│   │   ├── auth.go                # 登录、登出、密码修改、初始化向导
│   │   ├── config.go              # 配置读写、推荐、导入导出
│   │   ├── handler.go             # Handler 结构体定义与构造函数
│   │   ├── history.go             # 历史记录分页查询、日志查看
│   │   ├── merge.go               # 合并/清理任务 SSE 流式执行
│   │   └── status.go              # 系统状态、统计、主播列表
│   ├── middleware/
│   │   └── auth.go                # 认证中间件、频率限制、安全头
│   ├── model/
│   │   └── model.go               # 数据模型（HistoryRecord、ScheduleConfig 等）
│   ├── service/
│   │   ├── clean.go               # 智能清理服务
│   │   ├── history.go             # 历史记录持久化服务
│   │   ├── log.go                 # 日志文件写入与轮转
│   │   ├── merge.go               # 合并服务（主流程）
│   │   ├── scanner.go             # 文件扫描、分组、场次划分
│   │   └── scheduler.go           # 定时调度器
│   ├── ffmpeg/
│   │   ├── ffmpeg.go              # FFmpeg 进程管理（超时、进程组）
│   │   ├── concat.go              # TS 拼接合并（stream-copy）
│   │   ├── convert.go             # FLV → MP4 转换（via TS 中间格式）
│   │   ├── reencode.go            # 重编码 fallback（concat filter + 硬件加速）
│   │   └── validate.go            # 输出校验（元数据 + 多点解码）
│   └── utils/
│       ├── disk.go                # 磁盘用量查询（statfs）
│       ├── fileutil.go            # 文件工具（安全删除、命名规则）
│       ├── logutil.go             # 日志轮转、启动摘要
│       ├── notify.go              # Webhook 通知
│       ├── strings.go             # 字符串工具（白名单匹配、随机数）
│       └── video.go               # 视频文件名解析、时长查询
├── frontend/                       # Vue 3 前端（Vite 构建）
│   ├── src/
│   │   ├── api/                    # API 模块（7 个文件）
│   │   ├── views/                  # 页面组件（6 个视图）
│   │   ├── style/                  # Notion 风格设计系统
│   │   └── utils/                  # HTTP 客户端、SSE、格式化
│   └── package.json
├── templates/                      # 前端编译产物（Go 服务静态文件）
├── Dockerfile                      # 多阶段构建（Go 编译 + Alpine 运行时）
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
| GET | `/api/streamers` | 获取主播列表（文件数、磁盘占用） |
| GET | `/api/streamers/:name/files` | 获取指定主播的文件列表 |
| GET | `/api/files/:name` | 获取指定主播的文件列表（别名） |

### 任务执行（SSE 流式输出）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/merge` | 执行合并（可指定主播，SSE 流式返回进度） |
| POST | `/api/merge/manual` | 手动合并指定文件 |
| POST | `/api/merge/retry` | 重试合并 |
| POST | `/api/clean` | 执行清理（可指定主播，SSE 流式返回进度） |
| GET | `/api/clean/estimate` | 预估可清理文件数和大小 |
| POST | `/api/clean/emergency` | 紧急清理（可自定义目标阈值） |
| GET | `/api/run/:task` | SSE 执行任务（`merge` 或 `clean`） |

### 历史与日志

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/history` | 分页查询历史记录（支持 `task` 过滤） |
| GET | `/api/history/export` | 导出全部历史记录 |
| GET | `/api/logs/list/:task` | 获取日志文件列表 |
| GET | `/api/logs/content/:task` | 获取日志文件内容（最近 200 行） |

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

### 配置文件（config.json）

配置文件位于 `{LOG_DIR}/config.json`，首次初始化时自动生成。环境变量优先于配置文件。注意：`PASSWORD` 和 `SECRET_KEY` 使用 `json:"-"` 标记，不写入 config.json；密码单独持久化到 `{LOG_DIR}/.credentials.json`。

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
cd .. && go build -o bililive-helper ./cmd/server

# Docker 一键构建
docker compose up -d --build
```

### 运行测试

```bash
go test ./... -v -count=1
```

### 日志格式

日志存储在 `{LOG_DIR}/{task}_log/` 目录下，每天自动轮转，保留 30 天。

```
═══════════════════════════════════════════
[18:08:45] ▶ 开始 [全局] 合并
[七七7] → 发现 1 场，3 个片段
[七七7] ⚙ 合并 001.flv + 003.flv (1.6 GB)
[七七7] ✅ 合并成功 → 七七7-合并版.flv
[招财静宝] → 单文件，跳过
[失眠熊] → 已合并，跳过
[18:08:49] ⏹ 结束 · 扫描 12 个主播，合并 1 场次 (1.6 GB)
═══════════════════════════════════════════
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

### Q: Docker 容器内 FFmpeg 硬件加速不生效

当前仅支持 Rockchip 平台的 `h264_rkmpp` 硬件编码器，系统会自动检测并降级到 `libx264`。如需其他硬件加速支持，需修改 `internal/ffmpeg/reencode.go` 中的编码器检测逻辑。

---

## 许可证

[MIT License](LICENSE)
