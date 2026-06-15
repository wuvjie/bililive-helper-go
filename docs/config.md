# 配置参考

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `PASSWORD` | 登录密码（可选，优先通过初始化向导设置） | 首次启动自动生成 |
| `SECRET_KEY` | Session 签名密钥 | 首次启动自动生成 |
| `TARGET_DIR` | 录制视频目录 | `/vol2/1000/video/bililive-go/抖音` |
| `LOG_DIR` | 日志和配置存储目录 | `/vol1/1000/docker/bililive-helper-go` |
| `TZ` | 时区 | `Asia/Shanghai` |
| `COOKIE_SECURE` | Session cookie Secure 标志（HTTPS 时设为 true） | `false` |
| `WEBHOOK_URL` | 任务完成通知（Bark/Slack/飞书等） | 空 |
| `API_TOKEN` | API Token（Bearer 认证） | 空 |
| `PORT` | 监听端口 | `5000` |

## config.json

位于 `{LOG_DIR}/config.json`，首次初始化时自动生成。环境变量优先于配置文件。

`PASSWORD` 和 `SECRET_KEY` 不写入 config.json，密码单独存储在 `{LOG_DIR}/.credentials.json`（权限 0600）。

| 字段 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| `TARGET_DIR` | string | 录制视频目录 | `/vol2/1000/video/bililive-go/抖音` |
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
| `BACKUP_START_HOUR` | int | 静默时段开始小时 | 4 |
| `BACKUP_START_MINUTE` | int | 静默时段开始分钟 | 0 |
| `BACKUP_END_HOUR` | int | 静默时段结束小时 | 12 |
| `BACKUP_END_MINUTE` | int | 静默时段结束分钟 | 0 |
| `PORT` | int | 监听端口 | 5000 |
| `SESSION_VERSION` | int | Session 版本号（改密时自动递增） | 0 |
