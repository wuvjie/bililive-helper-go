# Bililive Helper Go

B站/抖音直播录像自动管理工具。自动合并录制片段、自动清理旧文件、Web 控制台管理。

## 功能概览

- **自动合并**：FLV→TS→MP4 转换合并（全程 -c copy，零重编码，RK3568 无压力）
- **自动清理**：磁盘超阈值时自动删除旧文件（白名单 + 安全期保护）
- **Web 控制台**：Vue 3 前端，实时进度、日志查看、历史记录
- **定时调度**：按间隔自动执行合并/清理，支持静默时段
- **硬件加速**：RK3568 MPP 硬件编码（重编码 fallback 时自动检测）
- **数据安全**：原子写入、进程组强杀、Context 取消链、fsync 刷盘

---

## 架构

```
cmd/server/main.go          启动入口（优雅停机 + 启动自检）
├── internal/config/         配置管理（原子写入 + 快照 + 事务回滚）
├── internal/service/
│   ├── merge.go            合并核心（TS转换 → 拼接 → MP4）
│   ├── clean.go            清理核心（阈值触发 + 白名单 + 安全期）
│   ├── scheduler.go        定时调度（Context 可取消 + 冷启动）
│   ├── history.go          操作历史（原子写入 + 自动清理）
│   └── log.go              日志轮转（每天归档 + 30天清理）
├── internal/handler/        HTTP 处理器（SSE 实时进度 + 防 XSS）
├── internal/middleware/     认证（防时序攻击）+ 安全头
├── internal/utils/          工具函数（文件解析、磁盘检测、ffmpeg、密码盐）
└── templates/index.html     Vue 3 前端（零构建工具 SPA）
```

---

## 合并模块（merge.go）

### 核心流程

```
Phase 0: 前置检查
  ├─ 磁盘空间 < 源文件×3+2GB → 跳过
  ├─ 静默时段 → 跳过
  └─ 目录不存在 → 报错退出

Phase 1: 扫描分类
  遍历每个主播文件夹
  ├─ 获取可解析视频（排除非视频、已合并、格式不对）
  ├─ 按 key 分组（主播名+标题）
  ├─ 按 mtime 升序排序（实际录制顺序）
  ├─ 按 gap(mtime-vs-dt) 分割批次（gap > 20分钟 = 不同场次）
  └─ 对每个批次判断：
       ├─ 输出已存在 + probe通过 → 删原片，跳过
       ├─ 单个 FLV → 检查录制中/安全锁 → 加入转换队列
       ├─ 不足 2 文件 → 跳过
       ├─ 录制中 → 跳过
       ├─ 文件太新（安全锁）→ 跳过
       ├─ 文件大小 < 1MB 或时长 < 5s → 跳过（损坏文件）
       └─ → 加入合并队列

Phase 2: FLV→MP4 转换（单文件）
  ├─ 检查文件未被占用 + 磁盘空间
  ├─ ffmpeg -c copy → .tmp.mp4
  ├─ 校验（大小 + probe）
  ├─ 通过 → 重命名 .tmp → .mp4 → 删 FLV
  └─ 失败 → 删 .tmp → 保留 FLV ✓

Phase 3: 多文件合并（TS 方案）
  ├─ 检查文件未被占用 + 磁盘空间
  ├─ FLV/MP4 → TS（-c copy + bsf:v h264_mp4toannexb）
  ├─ concat:1.ts|2.ts → 合并版.mp4
  ├─ 校验（probe + 10秒解码测试 + 大小）
  ├─ FLV→MP4 转换（如输入是 FLV）
  └─ 删所有原始文件（只在全部验证通过后）

  失败 fallback → concat filter 重编码（h264_rkmpp/libx264）

数据安全保证：
  任何环节失败 → 原始文件 100% 保留
  所有耗时操作响应 Context 取消（优雅关机）
  进程组强杀（syscall.Kill(-pid)）防止僵尸进程
```

---

## 清理模块（clean.go）

```
Phase 0: 前置检查
  ├─ 静默时段 → 跳过
  ├─ 全局模式：磁盘 < 触发阈值(85%) → 跳过
  └─ 指定主播：磁盘 > 95% → 报错

Phase 1: 计算需释放空间
  needToFree = 已用 - (总量 × 目标阈值70%)

Phase 2: 收集候选文件
  ├─ 获取所有视频（包括已合并的）
  ├─ 按 mtime 排序，保留最新 min_keep(3) 个
  ├─ 白名单过滤（不分大小写）
  └─ 安全期过滤（hours/days 模式）

Phase 3: 排序 + 删除
  ├─ 已合并文件优先删除（保护原始分片）
  ├─ 删除前检查文件存在 + 是否被占用
  ├─ 停止条件：达单次上限(20) 或 释放空间达标
  └─ 每个删除写入日志
```

---

## 调度模块（scheduler.go）

- Context 可取消：重启时自动终止 FFmpeg 进程
- 冷启动：重启后立即触发首次任务（lastRun 初始化为 now - interval）
- 并发控制：同一任务不重复执行
- Per-streamer 锁：防止同一主播操作冲突
- 锁超时保护：4 小时自动释放

---

## 配置模块（config.go）

- 原子写入：.tmp + rename，崩溃安全
- 事务回滚：Apply 失败自动回滚（深拷贝白名单切片）
- 快照模式：并发读安全
- 环境变量覆盖：TARGET_DIR、PASSWORD、LOG_DIR、SECRET_KEY

---

## 安全特性

- SHA-256 密码哈希（随机盐值）
- API Token constant-time 比较（防时序攻击）
- 登录速率限制（5次/5分钟）
- Session HttpOnly + SameSite=Lax
- 路径穿越防护
- XSS 防护（日志渲染前转义）
- 历史记录原子写入（防断电丢失）

---

## 日志

- 路径：`{LOG_DIR}/{task}_log/{task}_videos.log`
- 轮转：每天归档 + 30天清理
- 每次执行用 `═══` 分隔
- 每个文件夹一行结果

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

## 前端

- Vue 3 Composition API（零构建工具）
- SSE 流式进度 + Fetch POST 流式读取
- XSS 防护（日志渲染前转义）
- 防抖自动滚动
- 配置实时校验（Clamping）
- Vue 加载失败兜底

---

## 部署

```bash
docker compose up -d --build
```

访问 `http://NAS_IP:5689`

### 环境变量

| 变量 | 说明 |
|------|------|
| PASSWORD | 登录密码 |
| TARGET_DIR | 视频目录 |
| LOG_DIR | 日志目录 |
| SECRET_KEY | Session 密钥 |
| TZ | 时区（Asia/Shanghai） |
| WEBHOOK_URL | 通知 webhook（可选） |

### 资源限制（RK3568）

```yaml
deploy:
  resources:
    limits:
      cpus: '3.0'    # 4核留1核给录制
      memory: 4G     # 8G留4G给录制
```

## 技术栈

- **后端**: Go 1.23、Gin、zap
- **前端**: Vue 3（CDN）、原生 CSS
- **视频处理**: FFmpeg（FLV→TS→MP4 + h264_rkmpp 硬件加速）
- **部署**: Docker / Docker Compose

## 许可证

MIT License
