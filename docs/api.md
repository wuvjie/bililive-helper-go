# API 接口文档

所有 `/api/*` 路由均需认证（公开接口除外）。

## 认证与系统初始化

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/login` | 登录，成功后在 Cookie 中设置 session token（带连续失败冷却保护） |
| POST | `/api/logout` | 登出，清除 session Cookie 并使会话失效 |
| GET | `/api/auth/check` | 检查当前 session 是否有效 |
| POST | `/api/auth/change-password` | 修改管理员登录密码（自动递增 session 版本） |
| GET | `/api/setup/status` | 检查是否首次启动（未设置密码） |
| POST | `/api/setup/init` | 首次启动设置初始密码 |

## 系统状态与健康检查

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/health` | 容器健康检查（返回系统运行态） |
| GET | `/api/status` | 简要状态（磁盘使用率、主播数） |
| GET | `/api/status/detail` | 详细状态（磁盘、待合并文件、调度状态） |
| GET | `/api/stats` | 统计数据（今日/本月合并清理、7 天趋势、日均产出与余量预估） |
| GET | `/api/setup/check` | 系统诊断（FFmpeg、目录权限、磁盘空间） |

## 配置管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/config` | 获取当前配置（不含敏感字段） |
| POST | `/api/config` | 保存配置（部分更新） |
| GET | `/api/config/defaults` | 获取默认配置值 |
| GET | `/api/config/export` | 导出纯净配置备份（含全局策略、调度参数、主播规则） |
| POST | `/api/config/import` | 导入配置并覆盖恢复 |

## 定时调度

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/schedule` | 获取调度状态与静默时段 |
| POST | `/api/schedule` | 保存调度配置 |
| POST | `/api/schedule/run/:task` | 手动触发任务（`merge` 或 `clean`） |

## 主播与文件管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/streamers` | 主播列表（文件数、磁盘占用、最新视频时间、规则信息） |
| GET | `/api/streamers/:name/files` | 指定主播的文件列表（按修改时间倒序） |
| DELETE | `/api/streamers/:name/files` | 删除指定主播的单个录像文件（级联清理伴生弹幕与封面） |
| GET | `/api/streamers/:name/play` | 在线视频点播预览（支持 HTTP Range 拖动播放） |
| GET | `/api/streamers/:name/rule` | 获取指定主播的个性化规则（起止裁剪与免删保护） |
| POST | `/api/streamers/:name/rule` | 保存指定主播的个性化规则 |

## 任务执行（SSE 流式输出）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/merge` | 执行合并（可指定主播） |
| POST | `/api/merge/manual` | 手动合并指定文件 |
| POST | `/api/clean` | 执行清理（可指定主播） |

## 历史归档与操作日志

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/history` | 分页查询（支持 `task`、`streamer` 过滤） |
| DELETE | `/api/history/:id` | 删除单条历史任务记录 |
| POST | `/api/history/clear` | 清空全部历史任务记录 |
| GET | `/api/history/export` | 导出全部历史记录 |
| GET | `/api/logs/content/:task` | 操作日志内容（`log_id` 参数，返回最近 200 行或 raw 纯文本） |
