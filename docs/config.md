# 配置参考

## 目录挂载规范（Docker Compose / Docker CLI / WSL 部署）

为了彻底解耦宿主机环境与容器配置，容器内部采用固定标准路径：

| 容器内路径 | 说明 | 宿主机推荐挂载路径 |
|------------|------|-------------------|
| `/recordings` | 视频录制根目录（包含各主播子目录） | `bililive-go` 的录制输出根目录 |
| `/data` | 持久化数据目录（配置文件、日志、会话凭据） | 宿主机的专用数据目录（如 `/docker/bililive-helper-go` 或 `D:\docker\bililive-helper-go`） |

### 1. Docker Compose 挂载配置
在宿主机根目录下的 `.env` 文件中配置映射：
```env
RECORDINGS_DIR=/path/to/recordings
DATA_DIR=/path/to/data
PORT=5689
```

### 2. Docker CLI (`docker run`) 挂载参数
```bash
-v /path/to/recordings:/recordings \
-v /path/to/data:/data \
-e TARGET_DIR=/recordings \
-e LOG_DIR=/data
```

### 3. WSL Container (`wslc run`) 挂载参数
```powershell
-v D:\video\recordings:/recordings `
-v D:\docker\bililive-helper-go:/data `
-e TARGET_DIR=/recordings `
-e LOG_DIR=/data
```

---

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `TARGET_DIR` | 录制视频目录（容器内映射目录） | `/recordings` |
| `LOG_DIR` | 日志和配置存储目录（容器内映射目录） | `/data` |
| `TZ` | 时区 | `Asia/Shanghai` |
| `PORT` | 内部服务监听端口 | `5000` |
| `PASSWORD` | 登录密码（可选，优先通过初始化向导设置） | 首次启动自动生成 |
| `SECRET_KEY` | Session 签名密钥 | 首次启动自动生成 |
| `COOKIE_SECURE` | Session cookie Secure 标志（HTTPS 时设为 true） | `false` |
| `WEBHOOK_URL` | 任务完成通知（Bark/Slack/飞书等兼容 Webhook） | 空 |
| `API_TOKEN` | API Token（第三方程序 Bearer 认证） | 空 |

---

## config.json

位于 `{LOG_DIR}/config.json`（即宿主机挂载的 `DATA_DIR` 目录下），首次初始化时自动生成。环境变量优先级高于配置文件。

> 注：`PASSWORD` 和 `SECRET_KEY` 不写入 `config.json`，密码单独加密存储在 `{LOG_DIR}/.credentials.json`（文件权限 0600）。

| 字段 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| `TARGET_DIR` | string | 录制视频目录 | `/recordings` |
| `TRIGGER_THRESHOLD` | float | 清理触发阈值（磁盘使用率 %） | 80 |
| `TARGET_THRESHOLD` | float | 清理目标阈值（磁盘使用率 %） | 60 |
| `MIN_KEEP_PER_STREAMER` | int | 每主播最少保留文件数 | 3 |
| `SAFE_AGE_MINUTES` | int | 安全期（小时模式下的分钟数） | 120 |
| `GAP_MINUTES` | int | 场次间隔（分钟） | 30 |
| `MERGE_AGE_MINUTES` | int | 合并等待时间（分钟） | 30 |
| `WHITELIST_KEYWORDS` | []string | 白名单关键词 | `["留存","纪念","高能","生日","勿删"]` |
| `SAFE_MODE` | string | 安全期模式：`hours` 或 `days` | `hours` |
| `SAFE_DAYS` | int | 安全期天数（仅 days 模式） | 1 |
| `MAX_DELETE_PER_RUN` | int | 单次清理最大删除数 | 10 |
| `BACKUP_START_HOUR` | int | 静默时段开始小时（起止均为 0 表示默认关闭） | 0 |
| `BACKUP_START_MINUTE` | int | 静默时段开始分钟 | 0 |
| `BACKUP_END_HOUR` | int | 静默时段结束小时（起止均为 0 表示默认关闭） | 0 |
| `BACKUP_END_MINUTE` | int | 静默时段结束分钟 | 0 |
| `PORT` | int | 监听端口 | 5000 |
| `SESSION_VERSION` | int | Session 版本号（改密时自动递增） | 0 |
| `STREAMER_RULES` | object | 主播个性化规则（时段裁剪、独立保底数、免删保护） | `{}` |
