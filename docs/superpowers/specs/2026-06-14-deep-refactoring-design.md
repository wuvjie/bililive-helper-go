# bililive-helper-go 深度重构设计文档

> 日期：2026-06-14 | 方案：C — 深度重构 | 外部行为不变

## 目标

1. 消除跨文件重复代码（目录扫描 7 处、原子写入 4 处、CSS 类 6+ 处）
2. 核心业务逻辑可测试（通过接口抽象实现 mock）
3. 理清职责边界（业务逻辑归 service、handler 只做 HTTP 适配）
4. 拆分超大函数和 God Component
5. 前后端补全工程化基础设施（ESLint、Vitest、Go 单元测试）

## 约束

- 外部 API 端点路径、参数格式不变
- config.json / .credentials.json / schedule.json 格式不变
- Docker 部署方式和环境变量名不变
- 数据库不在本次范围内（保留文件存储）

---

## 第一阶段：后端基础设施层

### 1.1 新建 `internal/fsutil` 包 — 公共文件操作

当前状态：
- `AtomicWrite` 模式在 config.go、scheduler.go、history.go 中重复 4 次
- 目录扫描（遍历 TARGET_DIR 下的主播子目录）在 7+ 处独立实现
- `SafeUnlink`（带重试的删除）和 `ValidatePath`（防遍历攻击）散落在 utils 包

重构：
```
internal/fsutil/
  atomic.go     — AtomicWriteFile(path, data, perm) error
  scan.go       — ScanStreamerDirs(root) ([]StreamerDir, error)
  safeio.go     — SafeUnlink(path) error, SafeRemoveAll(path) error
  path.go       — ValidatePath(root, target) error
```

- `AtomicWriteFile`：write tmp → fsync → rename，一处实现，四处调用
- `ScanStreamerDirs`：统一目录扫描逻辑，返回标准化的 `StreamerDir{Name, Path, Files []os.DirEntry}`
- 所有现存的重复调用点全部迁移过来，删除原位置的重复代码

### 1.2 新建 `internal/taskctx` 包 — 统一任务上下文

当前状态：`MergeService.Run()` 和 `CleanService.Run()` 开头有 7 行完全相同的样板代码（snapshot config → create OpLogger → check backup window → check path → format tag → setup onProgress）。

重构：
```go
type TaskContext struct {
    Cfg      config.ConfigDTO   // 快照，不持有锁
    Logger   *zap.Logger
    OpLog    *OpLogger          // 可选，nil 时只走 SSE
    Progress func(string)       // SSE 推送
    Ctx      context.Context
    Cancel   context.CancelFunc
}

func New(cfg *config.Config, logger *zap.Logger, taskType string, progress func(string)) *TaskContext
```

- `MergeService.Run()` 和 `CleanService.Run()` 开头样板代码变成 `taskctx.New(...)` 一个调用

### 1.3 统一导出工具

当前状态：Settings 和 History 视图的 `handleExport()` 各自实现 JSON Blob 下载。

重构：在 `internal/handler` 中提取 `respondJSONFile(c *gin.Context, filename string, data any)` 辅助函数。

---

## 第二阶段：后端 Service 层重构

### 2.1 引入接口抽象 — `internal/service/interfaces.go`（新文件）

当前状态：`SchedulerService` 持有 `*MergeService` 和 `*CleanService` 具体类型指针，`runTask()` 用硬编码 switch 分发。

重构：
```go
// TaskRunner 是所有可调度任务的通用接口
type TaskRunner interface {
    Name() string
    Run(ctx context.Context, progress func(string)) (*TaskResult, error)
}

type TaskResult struct {
    Processed int
    Duration  time.Duration
    Details   string
}
```

- `MergeService` 和 `CleanService` 各自实现 `TaskRunner`
- `SchedulerService` 持有 `[]TaskRunner` 切片，用注册模式替代 switch
- 添加新任务类型只需实现接口并注册，无需修改 scheduler 代码（Open/Closed Principle）

### 2.2 拆分 MergeService 大函数

当前状态：
- `Run()` 172 行，混合扫描、磁盘检查、FLV 转换循环、合并循环、结果聚合
- `doMerge()` 148 行，混合输出命名、磁盘检查、TS 转换、拼接、验证、同步、时间戳保留、源文件删除

重构目标（每个函数 <50 行）：
```
MergeService.Run() 拆分为:
  — scanAndClassify()     → 扫描文件并分类（需要 FLV 转换的 vs 直接合并的）
  — convertFlvFiles()     → FLV → MP4 转换循环
  — mergeTasks()          → 合并任务循环
  — summarize()           → 结果汇总和历史记录

MergeService.doMerge() 拆分为:
  — prepareOutput()       → 输出路径命名和目录创建
  — executePipeline()     → TS 转换 + 拼接（或 reencode fallback）
  — validateAndFinalize() → 验证 + fsync + 保留时间戳 + 删除源文件
```

### 2.3 ffmpeg 包接口化

当前状态：service 层直接调用 `ffmpeg.Run()`、`ffmpeg.ConcatTS()` 等具体函数，无法 mock。

重构：
```go
// internal/ffmpeg/ffmpeg.go 中新增
type Executor interface {
    Run(ctx context.Context, args []string, opts RunOptions) ([]byte, error)
    ConcatTS(ctx context.Context, tsFiles []string, output string, onProgress func(float64)) error
    ConvertToMP4(ctx context.Context, input, output string) error
    Reencode(ctx context.Context, inputs []string, output string, onProgress func(float64)) error
    ValidateOutput(ctx context.Context, path string) error
    ProbeDuration(ctx context.Context, path string) (float64, error)
}

// 默认实现包装现有函数
type DefaultExecutor struct{}

// MergeService 通过构造函数注入 Executor 接口
func NewMergeService(cfg, logger, history, executor) *MergeService
```

- 单元测试时注入 mock executor，不再依赖实际 ffmpeg
- 现有的 `GetVideoDuration`、`IsVideoHealthy` 从 `utils/video.go` 迁移到 `ffmpeg` 包（它们本来就是 ffprobe 操作）

### 2.4 CleanService 内部优化

当前状态：
- `collectCandidates` 使用 `*[]candidateFile` 指针传参 + append 副作用
- `deleteFiles` 75 行混合快照、删除循环、大小比较、进度报告

重构：
- `collectCandidates` 改为返回 `[]candidateFile`，不用指针
- `deleteFiles` 拆分为 `snapshotSizes()` + `filterWritableFiles()` + `deleteWithProgress()`

### 2.5 业务逻辑移出 Handler

当前状态：
- `handler/merge.go` 中 `CleanEstimate`（45 行目录扫描+过滤）是纯业务逻辑
- `handler/merge.go` 中 `SetupCheck`（70 行系统诊断）是纯业务逻辑
- `handler/config.go` 中 `RecommendConfig` 和 `analyzeContent` 是纯业务逻辑

重构：
- `CleanEstimate` 逻辑移入 `CleanService.Estimate()` 方法
- `SetupCheck` 逻辑移入新 `DiagnosticsService` 或 `HealthService`
- `RecommendConfig` 和 `analyzeContent` 逻辑移入 `ConfigService`（从 config 包提升为 service）
- Handler 只保留 HTTP 解析 → 调用 service → 格式化响应

---

## 第三阶段：后端 Config 系统重构

### 3.1 ConfigDTO 自动化

当前状态：`Config` 和 `ConfigDTO` 手动镜像字段，`ApplyFromMap` 用 `map[string]interface{}`，`DiffDTO` 是 50 行逐字段比较。

重构方案：利用 struct tag + JSON 序列化/反序列化实现自动化：

```go
type Config struct {
    TargetDir       string  `json:"target_dir" env:"TARGET_DIR" default:""`
    MergeInterval   int     `json:"merge_interval" validate:"min=60,max=1440" default:"360"`
    // ...
}

// DTO 通过 json tag 自动从 Config 生成，排除 password/secret_key
func (c *Config) ToDTO() ConfigDTO {
    // 序列化 → 排除敏感字段 → 反序列化为 DTO
}

// Diff 通过反射对比两个 DTO 的差异
func DiffDTO(old, new ConfigDTO) []FieldChange

// ApplyFromMap 直接反序列化为 Config struct（类型安全）
func (c *Config) ApplyFromTyped(update ConfigUpdate) error
```

- 消除 `map[string]interface{}` 的类型不安全问题
- 新增配置字段只需在 struct 上加 tag，无需修改 DTO/diff/validate 代码

### 3.2 配置验证统一

当前状态：`Validate()` 函数中每个字段的手动 range check。

重构：用 struct tag `validate:"min=60,max=1440"` 驱动验证，可以引入轻量级验证库（如 `go-playground/validator`）或自行实现简单的 tag 解析。考虑到项目依赖极简的风格，建议自行实现一个 ~80 行的 tag 解析器。

---

## 第四阶段：后端 Handler 层重构

### 4.1 统一响应格式

当前状态：`gin.H` 散落各处，类型不安全。

重构：
```go
// internal/handler/response.go
type APIResponse struct {
    Code    int    `json:"code"`
    Message string `json:"message,omitempty"`
    Data    any    `json:"data,omitempty"`
}

func ok(c *gin.Context, data any)
func okMsg(c *gin.Context, msg string)
func fail(c *gin.Context, httpCode int, msg string)
func failWithCode(c *gin.Context, httpCode int, bizCode int, msg string)
```

- 所有 handler 统一使用这些函数，不再直接写 `c.JSON(200, gin.H{...})`

### 4.2 统一错误处理中间件

```go
// Recovery 中间件 + 统一 panic → 500 响应
// 业务错误类型：
type BizError struct {
    HTTPCode int
    BizCode  int
    Message  string
}

func ErrorHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()
        if len(c.Errors) > 0 {
            // 统一处理 BizError → JSON 响应
        }
    }
}
```

### 4.3 路由注册拆分

当前状态：main.go 中 40+ 条路由注册在一个平铺的代码块里。

重构：
```go
// internal/handler/routes.go
func RegisterRoutes(r *gin.Engine, h *Handler, limiter *RateLimiter) {
    public := r.Group("/api")
    { /* login, health, setup */ }

    auth := r.Group("/api")
    auth.Use(AuthRequired(), limiter.Limit())
    { /* 所有需要认证的路由按功能分组 */ }
}
```

- main.go 中只需一行 `handler.RegisterRoutes(r, h, limiter)`

### 4.4 SSE 辅助函数提取

当前状态：`runSSE()` 定义在 merge handler 中，但被 RunClean 和 EmergencyClean 也使用。

重构：移入 `internal/handler/sse.go` 作为共享辅助函数。

---

## 第五阶段：后端 main.go 重构

### 5.1 拆分为可测试结构

当前状态：`main()` 函数 160 行，无法测试。

重构：
```go
// cmd/server/main.go
func main() {
    app, err := NewApp()
    if err != nil { log.Fatal(err) }
    if err := app.Run(); err != nil { log.Fatal(err) }
}

// internal/app/app.go（新包）
type App struct {
    server   *http.Server
    scheduler *service.SchedulerService
    logger   *zap.Logger
}

func NewApp() (*App, error)    // 所有初始化逻辑
func (a *App) Run() error      // 启动 + graceful shutdown
func (a *App) Shutdown() error  // 可测试的关闭
```

### 5.2 LogStartup 参数对象化

当前状态：`utils.LogStartup` 接受 13 个位置参数。

重构：
```go
type StartupInfo struct {
    Port        int
    TargetDir   string
    LogDir      string
    FFmpegPath  string
    SchedulerOn bool
    // ...
}
func LogStartup(logger *zap.Logger, info StartupInfo)
```

---

## 第六阶段：前端架构重构

### 6.1 Settings God Component 拆分

当前状态：`settings/index.vue` 1008 行，6 个 tab 全在一个文件里。

重构：
```
views/settings/
  index.vue              — 壳组件，只管 tab 切换（~80 行）
  tabs/
    GeneralTab.vue       — 通用配置表单（~120 行）
    StorageTab.vue       — 存储管理 + 紧急清理（~180 行）
    ScheduleTab.vue      — 定时任务配置（~120 行）
    DiagnosticsTab.vue   — 系统诊断（~80 行）
    RecommendTab.vue     — 智能推荐（~150 行）
    BackupTab.vue        — 配置备份恢复（~80 行）
  composables/
    useSettings.ts       — 共享的配置加载、保存、dirty tracking 逻辑
```

### 6.2 提取 Composables 消除重复

当前状态：`onMounted` + `onActivated` 重复 fetch 逻辑在 3 个 view 中各复制粘贴了一遍。

重构：
```typescript
// composables/useFetchOnActivate.ts
export function useFetchOnActivate(fetchFn: () => Promise<void>) {
  onMounted(fetchFn)
  onActivated(fetchFn)
}

// composables/useStreamerData.ts — 多个 view 共享的主播数据
export function useStreamerData() {
  // 单例模式：多个组件调用只发一次请求
}

// composables/useDiskStatus.ts — layout + dashboard 共享
export function useDiskStatus() { ... }

// composables/useSchedule.ts — tasks + settings 共享
export function useSchedule() { ... }
```

### 6.3 Pinia Domain Stores

当前状态：Pinia store 只有 12 行（sidebar toggle），所有数据在各 view 独立 fetch。

重构：
```
store/modules/
  app.ts          — 保留 sidebar + 新增全局 loading/error 状态
  auth.ts         — 认证状态（从 router 模块迁移出来）
  streamers.ts    — 主播列表（多 view 共享，避免重复 fetch）
  schedule.ts     — 定时任务配置
  config.ts       — 系统配置
```

- 每个 store 提供 `fetch()` + 缓存，组件通过 `useXxxStore()` 消费
- 导航时不再触发重复 API 调用

### 6.4 Router Guard 状态迁移

当前状态：`router/index.ts` 中用模块级变量 `setupChecked`、`isAuthenticated` 管理认证状态，且 `markAuthenticated` 从 router 导出给 login/setup 视图使用。

重构：
- 认证状态移入 `store/modules/auth.ts`
- Router guard 读取 auth store，不再自己维护状态
- 消除 router 和 login/setup 视图之间的紧耦合

### 6.5 Layout 拆分

当前状态：`layout/index.vue` 316 行，包含磁盘告警横幅、密码修改对话框、侧边栏、导航栏。

重构：
```
layout/
  index.vue              — 主壳（~100 行）
  components/
    AppSidebar.vue       — 侧边栏
    AppNavbar.vue        — 顶部导航栏
    DiskAlert.vue        — 磁盘告警横幅
    PasswordDialog.vue   — 密码修改对话框
```

---

## 第七阶段：前端工程化

### 7.1 引入 ESLint + Prettier

```
# 安装
npm i -D eslint @vue/eslint-config-typescript @vue/eslint-config-prettier prettier

# .eslintrc.cjs 关键配置
extends: [
  'eslint:recommended',
  '@vue/eslint-config-typescript',
  '@vue/eslint-config-prettier'
]
rules:
  '@typescript-eslint/no-explicit-any': 'warn'
  no-console: 'warn'
```

- 配置 `lint-staged` + `husky` 在 pre-commit 自动格式化

### 7.2 CSS 设计系统

当前状态：`.mono-val`、`.card`、`.status-dot-sm` 等工具类在 6 个组件中重复定义。

重构：
```scss
// style/_tokens.scss — 设计 token（颜色、间距、字体）
$color-success: #448361;
$color-error: #e03131;
$color-muted: #888888;

// style/_components.scss — 共享组件类
.mono-val { font-family: var(--font-mono); ... }
.card { background: #fff; border-radius: 12px; ... }
.status-dot { width: 8px; height: 8px; border-radius: 50%; ... }
.dot-ok { background: $color-success; }
.dot-err { background: $color-error; }
```

- 各组件删除重复的 scoped 样式，改用全局工具类

### 7.3 TypeScript 类型加固

- `Record<string, any>` → 定义具体的 payload 类型（`ConfigUpdate`、`ScheduleUpdate` 等）
- `http.ts` 的 `get<T>()` / `post<T>()` 默认泛型从 `any` 改为 `unknown`
- 添加 `app.config.errorHandler` 全局错误边界
- SSE composable：`startSSE` 返回 `Promise<void>` 但增加 `onComplete` 回调
- 修正 `isMobile` 为响应式（监听 `resize` 事件或使用 `@vueuse/core` 的 `useMediaQuery`）

### 7.4 引入 Vitest

```
npm i -D vitest @vue/test-utils jsdom

# vitest.config.ts
export default defineConfig({
  test: { environment: 'jsdom' }
})
```

优先测试目标：
1. Composables（`useFetchOnActivate`、`useStreamerData`）
2. API 类型正确性
3. 格式化工具函数（`format.ts`）

---

## 第八阶段：后端测试补全

### 8.1 可测试性基础设施

在第二阶段引入接口抽象后，可以开始补测试：

```go
// internal/service/mock_test.go — 测试用 mock
type mockExecutor struct { ... }      // 实现 ffmpeg.Executor
type mockHistory struct { ... }       // 实现 HistoryRecorder 接口
type mockConfig struct { ... }        // 返回固定配置快照
```

### 8.2 优先测试目标

| 优先级 | 目标 | 原因 |
|--------|------|------|
| P0 | `MergeService` 扫描+分类逻辑 | 核心业务，已有部分测试（SortByFilename 等），扩展 |
| P0 | `CleanService` 候选文件选择逻辑 | 当前零测试，逻辑复杂 |
| P1 | `SchedulerService` 调度逻辑 | 接口化后可 mock 测试 |
| P1 | `config` 验证和 Apply 逻辑 | 当前零测试 |
| P2 | `handler` 层集成测试 | 依赖前面的接口抽象 |

---

## 执行顺序

```
阶段 1 — fsutil + taskctx（基础设施）         ← 无风险，纯提取
阶段 2 — service 接口化 + 函数拆分            ← 核心改动，需仔细验证
阶段 3 — config 系统重构                      ← 独立模块，影响面可控
阶段 4 — handler 层重构 + 路由拆分            ← 依赖阶段 2 的接口
阶段 5 — main.go 拆分                         ← 依赖阶段 4 的路由注册
阶段 6 — 前端组件拆分 + composables           ← 前后端可并行
阶段 7 — 前端工程化（ESLint/Vitest/CSS）      ← 依赖阶段 6
阶段 8 — 后端测试补全                         ← 依赖阶段 2-3 的接口
```

每个阶段完成后：`go build ./...` + 现有测试通过 + 手动冒烟测试。

---

## 文件变更概览

### 新增文件
```
internal/fsutil/atomic.go
internal/fsutil/scan.go
internal/fsutil/safeio.go
internal/fsutil/path.go
internal/taskctx/taskctx.go
internal/service/interfaces.go
internal/handler/response.go
internal/handler/routes.go
internal/handler/sse.go
internal/app/app.go
frontend/src/views/settings/tabs/*.vue (6 files)
frontend/src/views/settings/composables/useSettings.ts
frontend/src/composables/useFetchOnActivate.ts
frontend/src/composables/useStreamerData.ts
frontend/src/composables/useDiskStatus.ts
frontend/src/composables/useSchedule.ts
frontend/src/store/modules/auth.ts
frontend/src/store/modules/streamers.ts
frontend/src/store/modules/schedule.ts
frontend/src/store/modules/config.ts
frontend/src/layout/components/*.vue (4 files)
frontend/src/style/_tokens.scss
frontend/src/style/_components.scss
```

### 重构（大幅修改）文件
```
internal/config/config.go          — 拆分为 config.go + dto.go + validate.go
internal/service/merge.go          — 函数拆分 + 接口注入
internal/service/clean.go          — 函数拆分 + 接口注入
internal/service/scheduler.go      — 注册模式替代 switch
internal/handler/merge.go          — 业务逻辑移出 + typed response
internal/handler/config.go         — 业务逻辑移出 + typed response
internal/handler/status.go         — typed response
internal/handler/auth.go           — typed response
cmd/server/main.go                 — 瘦身为 ~20 行
frontend/src/views/settings/index.vue — 从 1008 行瘦身到 ~80 行
frontend/src/layout/index.vue      — 拆分为壳 + 子组件
frontend/src/utils/http.ts         — 类型加固
frontend/src/router/index.ts       — 认证状态迁移
```

### 删除文件（功能合并后）
```
internal/utils/video.go            — ffprobe 相关函数迁入 ffmpeg 包后删除
internal/utils/disk.go             — 迁入 fsutil 后删除
internal/utils/disk_windows.go     — 迁入 fsutil 后删除
```
