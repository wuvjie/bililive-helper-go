# API 接口文档

所有 `/api/*` 路由（除公开接口外）均需登录认证。认证方式支持 Session Cookie 和 Bearer Token。

## 公开接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/login` | 用户登录（JSON: `{password}`） |
| GET | `/api/health` | 健康检查 |
| GET | `/api/setup/status` | 查询是否为首次运行 |
| POST | `/api/setup/init` | 首次初始化（设置密码） |

## 认证

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/auth/check` | 检查登录状态 |
| POST | `/api/auth/change-password` | 修改密码 |

## 系统状态

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/status` | 简要状态（磁盘使用率、主播数） |
| GET | `/api/status/detail` | 详细状态（磁盘、待合并文件、调度状态） |
| GET | `/api/stats` | 统计数据（今日/本月合并清理、7 天趋势） |
| GET | `/api/setup/check` | 系统诊断（FFmpeg、目录权限、磁盘空间） |

## 配置管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/config` | 获取当前配置（不含敏感字段） |
| POST | `/api/config` | 保存配置（部分更新） |
| GET | `/api/config/recommend` | 智能推荐配置（基于磁盘和内容分析） |
| GET | `/api/config/defaults` | 获取默认配置值 |
| GET | `/api/config/export` | 导出完整配置（含调度、历史） |
| POST | `/api/config/import` | 导入配置 |

## 定时调度

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/schedule` | 获取调度状态 |
| POST | `/api/schedule` | 保存调度配置 |
| POST | `/api/schedule/run/:task` | 手动触发任务（`merge` 或 `clean`） |

## 主播与文件

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/streamers` | 主播列表（文件数、磁盘占用、最新视频时间） |
| GET | `/api/streamers/:name/files` | 指定主播的文件列表 |

## 任务执行（SSE 流式输出）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/merge` | 执行合并（可指定主播） |
| POST | `/api/merge/manual` | 手动合并指定文件 |
| POST | `/api/merge/retry` | 重试合并 |
| POST | `/api/clean` | 执行清理（可指定主播） |
| GET | `/api/clean/estimate` | 预估可清理文件数和大小 |
| POST | `/api/clean/emergency` | 紧急清理 |

## 历史与日志

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/history` | 分页查询（支持 `task`、`streamer` 过滤） |
| GET | `/api/history/export` | 导出全部历史记录 |
| GET | `/api/logs/content/:task` | 操作日志内容（`log_id` 参数，返回最近 200 行） |
