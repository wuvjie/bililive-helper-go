# bililive-helper-go 深度重构设计文档

> 版本：2.0 | 日期：2026-06-14 | 方案：C — 深度重构 + 功能增强 | 外部行为不变

---

## 一、目标

### 代码质量（解决"改不动"）
1. 消除跨文件重复代码（目录扫描 7 处、原子写入 4 处、CSS 类 6+ 处）
2. 核心业务逻辑可测试（通过接口抽象实现 mock）
3. 理清职责边界（业务逻辑归 service、handler 只做 HTTP 适配）
4. 拆分超大函数和 God Component（172 行 Run → 每个 <50 行，1008 行 Settings → 6 个组件）
5. 依赖升级到最新版本（Gin 1.12、Pinia 3、Vue Router 5、Vite 8、TypeScript 6）

### 功能增强（解决"不够好"）
6. 任务队列与并发控制（可配置并发上限，避免 20 个 ffmpeg 同时打满磁盘 I/O）
7. 文件健康预检（合并前 ffprobe 快检，跳过损坏文件，避免浪费几十分钟）
8. 事件驱动调度（目录变化检测替代纯固定间隔，下播后文件更快被处理）

### 工程化（解决"不敢改"）
9. 前后端补全工程化基础设施（ESLint、Vitest、Go 单元测试、静态分析）

---

## 二、约束

- 外部 API 端点路径、参数格式不变
- config.json / .credentials.json / schedule.json 格式不变（新增字段向后兼容）
- Docker 部署方式和环境变量名不变
- 保留文件存储（不引入数据库）
- 技术栈不变（Go + Gin + Vue 3 + Element Plus + Vite，已验证为最佳选择）

---

## 三、方法论

| 方法论 | 来源 | 在本项目中的应用 |
|--------|------|-----------------|
| 特征测试先行 | Michael Feathers | 重构前记录当前行为基线，重构后对比确保无回归 |
| 绞杀者模式 | Martin Fowler | 新旧代码短暂共存，逐个迁移验证后再删除旧代码 |
| Mikado 依赖图 | Ellnestam & Brolund | 阶段 2 内部画依赖图，叶子节点优先执行 |
| 先让改动变简单 | Kent Beck | 每阶段先创造条件，再做目标改动 |
| ADR 决策记录 | Michael Nygard | 每个架构决策记录 `docs/adr/`，记录"为什么" |
| 风险量化 | Fowler 债务象限 | golangci-lint + gocyclo 建立基线，数据驱动优先级 |
| API 契约验证 | Golden Master | 重构前记录 API 响应，重构后 diff 对比 |
| 小步提交 | Fowler | 叶子节点独立提交，代码库始终可编译可运行 |

### 安全网

```
Level 0 — 静态分析基线（golangci-lint + gocyclo + vue-tsc）
Level 1 — 特征测试（记录当前行为，重构后对比）
Level 2 — API 响应基线（Golden Master，前后端兼容性保障）
Level 3 — 单元测试（重构过程中逐步补充）
Level 4 — 手动冒烟测试（每个阶段完成后）
```

### 关键规则

- 每个叶子节点改动独立提交，代码库**始终可编译可运行**
- 每阶段完成后运行特征测试，确认行为未变
- 重构提交和功能提交**严格分开**
- 阶段 8 完成后对比 golangci-lint / gocyclo 输出，确认复杂度量化下降

---

## 四、阶段总览

```
阶段 -1 — 建立安全网         ← 最最优先，任何代码改动之前
阶段  0 — 依赖升级           ← 确保基于最新 API
阶段  1 — 后端基础设施层     ← 提取公共操作，绞杀者模式逐点迁移
阶段  2 — 后端 Service 层    ← 接口化 + 函数拆分 + 任务队列 + 事件驱动
阶段  3 — 后端 Config 系统   ← DTO 自动化 + 验证统一
阶段  4 — 后端 Handler 层    ← 统一响应 + 业务逻辑移出 + 路由拆分
阶段  5 — 后端 main.go       ← 拆分为可测试的 App 结构
阶段  6 — 前端架构重构       ← 组件拆分 + Composables + Pinia stores
阶段  7 — 前端工程化         ← ESLint + Vitest + CSS 设计系统 + 类型加固
阶段  8 — 测试补全 + 回归验证 ← 补单元测试 + 对比阶段 -1 基线
```

---

## 五、详细设计

---

### 阶段 -1：建立安全网（重构前必做）

> Michael Feathers: "Legacy code is code without tests."

在任何代码改动之前，建立行为基线。

#### -1.1 静态分析基线

```bash
golangci-lint run ./...               # Go 代码质量
gocyclo -over 15 ./internal/...       # 圈复杂度
cd frontend && npx vue-tsc --noEmit   # 前端类型检查
```

输出记录到 `tests/baseline/`，重构后对比。

#### -1.2 后端特征测试

记录当前行为（不验证"正确"），作为重构的安全网：

| 目标 | 测试内容 |
|------|---------|
| `MergeService.scanTasks` | 模拟目录结构，记录分类结果 |
| `MergeService.SortByFilename` | 已有测试，扩展边界用例 |
| `CleanService.collectCandidates` | 模拟文件，记录清理优先级 |
| `config.Validate` | 边界值，记录验证结果 |
| `config.Apply` | 原子写入行为 |
| `ParseFilename` | 已有测试，确认覆盖率 |

#### -1.3 API 响应基线

```bash
curl -s http://localhost:5000/api/health    > tests/baseline/api_health.json
curl -s http://localhost:5000/api/status    > tests/baseline/api_status.json
# ... 对所有端点重复
```

重构后 diff 对比，确保 API 契约不变。

#### -1.4 前端构建基线

```bash
npx vue-tsc --noEmit > tests/baseline/vue-tsc-before.txt 2>&1
npm run build 2>&1 | tee tests/baseline/build-before.txt
```

#### -1.5 ADR 记录

为首个架构决策创建 `docs/adr/0001-fsutil-package.md`。

**交付物**：`tests/baseline/` 目录 + 扩展后的 `*_test.go` + `docs/adr/` 目录

---

### 阶段 0：依赖升级

> 先升后改 — 避免重构时基于旧 API 写代码，升级后又要改。

#### 0.1 后端（零代码改动）

| 依赖 | 当前 → 目标 | 风险 |
|------|------------|------|
| Go | 1.25 → 1.26 | 🟢 无 |
| Gin | v1.9.1 → v1.12.0 | 🟢 无 |
| Zap | v1.26.0 → v1.28.0 | 🟢 无 |
| x/crypto | v0.52.0 → latest | 🟢 无 |
| sessions | v0.0.5 → v1.1.0 | 🟡 API 兼容，测 cookie |

```bash
go get github.com/gin-gonic/gin@v1.12.0 go.uber.org/zap@v1.28.0 golang.org/x/crypto@latest github.com/gin-contrib/sessions@v1.1.0
go mod tidy && go build ./... && go test ./...
```

#### 0.2 前端（分 4 步，每步验证）

**Step 1** — patch/minor（🟢 零风险）
```bash
npm update vue axios element-plus sass
npm run build
```

**Step 2** — Vite 6 → 8（🟡 改 vite.config.ts）
```diff
- rollupOptions: {
+ rolldownOptions: {
```
```bash
npm install vite@^8.0.0 @vitejs/plugin-vue@^6.0.0
npm run build
```

**Step 3** — TypeScript 5 → 6（🟡 改 tsconfig.json）
```diff
+ "rootDir": "./src",
  "paths": {
-   "@/*": ["src/*"]
+   "@/*": ["./src/*"]
  }
```
```bash
npm install typescript@^6.0.0
npx vue-tsc --noEmit
```

**Step 4** — Pinia 3 + Vue Router 5（🟡 改 guard 写法）
```diff
- router.beforeEach(async (to, _from, next) => {
+ router.beforeEach(async (to, _from) => {
-   next("/login");
+   return "/login";
```
```bash
npm install pinia@^3.0.0 vue-router@^5.0.0
npm run build
```

**交付物**：更新后的 go.mod + package.json + vite.config.ts + tsconfig.json + router/index.ts

---

### 阶段 1：后端基础设施层

> 绞杀者模式：新代码与旧代码短暂共存，逐个迁移验证后再删除旧代码。

#### 1.1 新建 `internal/fsutil` — 公共文件操作

```
internal/fsutil/
  atomic.go     — AtomicWriteFile(path, data, perm) error
  scan.go       — ScanStreamerDirs(root) ([]StreamerDir, error) + 文件健康预检
  safeio.go     — SafeUnlink(path) error
  path.go       — ValidatePath(root, target) error
```

**消除的重复**：
- `AtomicWriteFile` → 替代 config.go、scheduler.go、history.go 中 4 处重复的 write-tmp-fsync-rename
- `ScanStreamerDirs` → 替代 7+ 处独立实现的目录扫描

**文件健康预检**（功能增强）：
- 扫描时同步调 ffprobe 快检（文件大小 > 0 且可读取时长）
- 返回 `FileHealth` 标记：Healthy / Corrupt / Skippable
- 合并时自动跳过损坏文件，避免浪费几十分钟的 ffmpeg 处理

**迁移策略**：
1. 建 fsutil，在 scheduler.go 的原子写入处替换
2. 运行特征测试验证
3. 确认无误后替换下一处
4. 旧函数标记 `// DEPRECATED: use fsutil.XXX`
5. 全部迁移后删除旧代码

#### 1.2 新建 `internal/taskctx` — 统一任务上下文

```go
type TaskContext struct {
    Cfg      config.ConfigDTO   // 快照，不持有锁
    Logger   *zap.Logger
    OpLog    *OpLogger          // 可选
    Progress func(string)       // SSE 推送
    Ctx      context.Context
    Cancel   context.CancelFunc
}

func New(cfg *config.Config, logger *zap.Logger, taskType string, progress func(string)) *TaskContext
```

消除 MergeService.Run() 和 CleanService.Run() 开头 7 行重复样板代码。

#### 1.3 统一导出工具

提取 `respondJSONFile(c, filename, data)` 辅助函数，替代 Settings 和 History 中重复的 JSON Blob 下载逻辑。

**交付物**：`internal/fsutil/` + `internal/taskctx/` + 迁移标记后的旧代码 + 扩展后的测试

---

### 阶段 2：后端 Service 层重构

> Mikado 方法：画依赖图，叶子节点优先。

#### Mikado 依赖图

```
目标：可测试 + 函数 <50 行 + 任务队列 + 事件驱动
│
├── 叶子节点（先做）：
│   ├── 2.4a collectCandidates 返回值重构
│   ├── 2.4b deleteFiles 拆分
│   └── 2.2a ffprobe 函数迁入 ffmpeg 包
│
├── 中间节点：
│   ├── 2.3 ffmpeg.Executor 接口 + DefaultExecutor
│   ├── 2.1 TaskRunner 接口
│   ├── 2.1a TaskQueue 实现
│   ├── 2.1b DirWatcher 实现
│   └── 2.2b MergeService.Run() 拆分
│
└── 根节点：
    ├── 2.5 业务逻辑移出 Handler
    ├── MergeService/CleanService 实现 TaskRunner
    └── SchedulerService 改造（TaskQueue + DirWatcher + 兜底定时器）
```

#### 2.1 TaskRunner 接口

```go
type TaskRunner interface {
    Name() string
    Run(ctx context.Context, progress func(string)) (*TaskResult, error)
}
```

MergeService 和 CleanService 各自实现 TaskRunner。SchedulerService 用 `[]TaskRunner` 注册模式替代硬编码 switch。

#### 2.1a 任务队列（功能增强）

```go
type TaskQueue struct {
    sem   chan struct{}      // 并发信号量，如 make(chan struct{}, 3)
    queue chan TaskRequest   // 等待队列
}
```

- 用户配置 `max_concurrent_tasks`（默认 3）
- 超出并发数的任务排队等待
- 新增配置项向后兼容

#### 2.1b 事件驱动调度（功能增强）

```go
type DirWatcher struct {
    root     string
    interval time.Duration        // 轮询间隔（轻量，不用 fsnotify）
    onChange func(streamer string)
}
```

- 每分钟检查各主播目录修改时间
- N 分钟无变化 → 认为录制结束 → 提交合并任务
- 防抖时间可配置（防止录制中途短暂暂停误判）
- 固定间隔降级为兜底（每天凌晨跑一次确保无遗漏）

#### 2.2 拆分 MergeService 大函数

```
Run() 172 行 →
  scanAndClassify()      扫描 + 分类
  convertFlvFiles()      FLV → MP4 转换
  mergeTasks()           合并循环
  summarize()            结果汇总

doMerge() 148 行 →
  prepareOutput()        输出路径 + 目录创建
  executePipeline()      TS 转换 + 拼接（或 reencode fallback）
  validateAndFinalize()  验证 + fsync + 时间戳 + 删除源文件
```

#### 2.3 ffmpeg 接口化

```go
type Executor interface {
    Run(ctx, args, opts) ([]byte, error)
    ConcatTS(ctx, tsFiles, output, onProgress) error
    ConvertToMP4(ctx, input, output) error
    Reencode(ctx, inputs, output, onProgress) error
    ValidateOutput(ctx, path) error
    ProbeDuration(ctx, path) (float64, error)
}

type DefaultExecutor struct{}  // 包装现有函数
```

MergeService 通过构造函数注入 Executor 接口，测试时注入 mock。

#### 2.4 CleanService 优化

- `collectCandidates`：`*[]candidateFile` 指针 → 返回 `[]candidateFile`
- `deleteFiles` 75 行 → 拆分为 `snapshotSizes()` + `filterWritableFiles()` + `deleteWithProgress()`

#### 2.5 业务逻辑移出 Handler

| Handler 中的业务逻辑 | 移入 |
|---------------------|------|
| `CleanEstimate`（45 行扫描） | `CleanService.Estimate()` |
| `SetupCheck`（70 行诊断） | 新 `DiagnosticsService` |
| `RecommendConfig` + `analyzeContent` | `ConfigService` |

Handler 只保留：HTTP 解析 → 调用 service → 格式化响应。

**交付物**：`interfaces.go` + `taskqueue.go` + `dirwatcher.go` + 拆分后的 merge.go / clean.go / scheduler.go + 扩展后的测试

---

### 阶段 3：后端 Config 系统重构

#### 3.1 ConfigDTO 自动化

```go
type Config struct {
    TargetDir     string `json:"target_dir" env:"TARGET_DIR" default:""`
    MergeInterval int    `json:"merge_interval" validate:"min=60,max=1440" default:"360"`
    // ...
}
```

- `ToDTO()` — JSON 序列化 → 排除敏感字段 → 反序列化为 DTO
- `DiffDTO()` — 反射对比两个 DTO 的差异（替代 50 行逐字段比较）
- `ApplyFromTyped()` — 直接反序列化为 Config struct（替代 `map[string]interface{}`）

#### 3.2 验证统一

struct tag `validate:"min=60,max=1440"` 驱动验证。自行实现 ~80 行的 tag 解析器（保持依赖极简）。

**交付物**：config.go 拆分为 config.go + dto.go + validate.go

---

### 阶段 4：后端 Handler 层重构

#### 4.1 统一响应

```go
func ok(c *gin.Context, data any)
func okMsg(c *gin.Context, msg string)
func fail(c *gin.Context, httpCode int, msg string)
func failWithCode(c *gin.Context, httpCode int, bizCode int, msg string)
```

替代散落各处的 `c.JSON(200, gin.H{...})`。

#### 4.2 错误处理中间件

```go
type BizError struct { HTTPCode, BizCode int; Message string }
func ErrorHandler() gin.HandlerFunc  // Recovery + 统一 panic → 500
```

#### 4.3 路由注册拆分

```go
// internal/handler/routes.go
func RegisterRoutes(r *gin.Engine, h *Handler, limiter *RateLimiter) {
    public := r.Group("/api")  { /* login, health, setup */ }
    auth := r.Group("/api")    { auth.Use(AuthRequired(), limiter.Limit()); /* 认证路由 */ }
}
```

main.go 中只需 `handler.RegisterRoutes(r, h, limiter)`。

#### 4.4 SSE 辅助提取

`runSSE()` 从 merge handler 移入 `internal/handler/sse.go` 共享。

**交付物**：`response.go` + `routes.go` + `sse.go` + 重构后的 handler 文件

---

### 阶段 5：后端 main.go 重构

#### 5.1 拆分为 App 结构

```go
// cmd/server/main.go
func main() {
    app, err := NewApp()
    if err != nil { log.Fatal(err) }
    if err := app.Run(); err != nil { log.Fatal(err) }
}

// internal/app/app.go
type App struct { server *http.Server; scheduler *service.SchedulerService; logger *zap.Logger }
func NewApp() (*App, error)
func (a *App) Run() error
```

#### 5.2 LogStartup 参数化

```go
type StartupInfo struct { Port int; TargetDir string; LogDir string; ... }
func LogStartup(logger *zap.Logger, info StartupInfo)
```

替代当前 13 个位置参数。

**交付物**：`internal/app/app.go` + 瘦身后的 `cmd/server/main.go`（~20 行）

---

### 阶段 6：前端架构重构

#### 6.1 Settings 拆分（1008 行 → 6 组件）

```
views/settings/
  index.vue                    壳组件，tab 切换（~80 行）
  tabs/GeneralTab.vue          通用配置（~120 行）
  tabs/StorageTab.vue          存储管理 + 紧急清理（~180 行）
  tabs/ScheduleTab.vue         定时任务（~120 行）
  tabs/DiagnosticsTab.vue      系统诊断（~80 行）
  tabs/RecommendTab.vue        智能推荐（~150 行）
  tabs/BackupTab.vue           备份恢复（~80 行）
  composables/useSettings.ts   共享逻辑
```

#### 6.2 Composables 消除重复

```typescript
useFetchOnActivate(fetchFn)    // 替代 3 个 view 中 120 行 onMounted+onActivated 复制粘贴
useStreamerData()              // 多 view 共享主播数据（单例）
useDiskStatus()                // layout + dashboard 共享
useSchedule()                  // tasks + settings 共享
```

#### 6.3 Pinia Domain Stores

```
store/modules/
  app.ts          sidebar + 全局 loading/error
  auth.ts         认证状态（从 router 模块迁移）
  streamers.ts    主播列表（避免重复 fetch）
  schedule.ts     定时任务配置
  config.ts       系统配置
```

#### 6.4 Router Guard 迁移

认证状态从 router 模块变量 → `store/modules/auth.ts`。消除 router 与 login/setup 视图的紧耦合。

#### 6.5 Layout 拆分（316 行 → 4 组件）

```
layout/
  index.vue                    主壳（~100 行）
  components/AppSidebar.vue    侧边栏
  components/AppNavbar.vue     顶部导航栏
  components/DiskAlert.vue     磁盘告警
  components/PasswordDialog.vue 密码修改
```

**交付物**：6 个 tab 组件 + 4 个 layout 组件 + 4 个 composables + 4 个 Pinia stores

---

### 阶段 7：前端工程化

#### 7.1 ESLint + Prettier

```bash
npm i -D eslint @vue/eslint-config-typescript @vue/eslint-config-prettier prettier lint-staged husky
```

规则：`@typescript-eslint/no-explicit-any: warn`，`no-console: warn`。配置 lint-staged pre-commit 自动格式化。

#### 7.2 CSS 设计系统

```scss
// style/_tokens.scss — 颜色、间距、字体 token
// style/_components.scss — 共享工具类（.mono-val, .card, .status-dot 等）
```

各组件删除重复的 scoped 样式，改用全局类。

#### 7.3 TypeScript 类型加固

- `Record<string, any>` → 具体 payload 类型
- `http.ts` 默认泛型 `any` → `unknown`
- 添加 `app.config.errorHandler` 全局错误边界
- 修正 `isMobile` 为响应式

#### 7.4 Vitest

```bash
npm i -D vitest @vue/test-utils jsdom
```

优先测试：composables > 工具函数 > API 类型。

**交付物**：`.eslintrc.cjs` + `prettier.config.js` + `_tokens.scss` + `_components.scss` + `vitest.config.ts` + 测试文件

---

### 阶段 8：测试补全 + 回归验证

#### 8.1 后端单元测试

```go
type mockExecutor struct { ... }      // 实现 ffmpeg.Executor
type mockHistory struct { ... }       // 实现 HistoryRecorder
```

| 优先级 | 目标 | 原因 |
|--------|------|------|
| P0 | MergeService 扫描+分类 | 核心业务，扩展已有测试 |
| P0 | CleanService 候选选择 | 当前零测试，逻辑复杂 |
| P1 | SchedulerService 调度+队列 | 接口化后可 mock |
| P1 | config 验证和 Apply | 当前零测试 |
| P2 | handler 层集成测试 | 依赖接口抽象 |

#### 8.2 回归验证

```bash
# 对比阶段 -1 的基线
golangci-lint run ./...       # 对比 lint-before.txt
gocyclo -over 15 ./internal/  # 对比 cyclo-before.txt
diff tests/baseline/api_*.json tests/current/api_*.json  # API 契约不变
npx vue-tsc --noEmit          # 对比 vue-tsc-before.txt
npm run build                 # 对比 build-before.txt
```

**交付物**：扩展后的 `*_test.go` + 回归对比报告

---

## 六、文件变更总览

### 新增文件

```
internal/fsutil/atomic.go                  原子写入
internal/fsutil/scan.go                    目录扫描 + 文件健康预检
internal/fsutil/safeio.go                  安全删除
internal/fsutil/path.go                    路径校验
internal/taskctx/taskctx.go                统一任务上下文
internal/service/interfaces.go             TaskRunner 接口
internal/service/taskqueue.go              任务队列与并发控制
internal/service/dirwatcher.go             目录变化监听
internal/handler/response.go               统一响应
internal/handler/routes.go                 路由注册
internal/handler/sse.go                    SSE 辅助
internal/app/app.go                        应用生命周期
frontend/src/views/settings/tabs/*.vue     6 个 tab 组件
frontend/src/views/settings/composables/   useSettings
frontend/src/composables/*.ts              4 个 composables
frontend/src/store/modules/*.ts            4 个 Pinia stores
frontend/src/layout/components/*.vue       4 个 layout 子组件
frontend/src/style/_tokens.scss            设计 token
frontend/src/style/_components.scss        共享工具类
```

### 重构文件

```
internal/config/config.go              → config.go + dto.go + validate.go
internal/service/merge.go              函数拆分 + 接口注入
internal/service/clean.go              函数拆分 + 接口注入
internal/service/scheduler.go          注册模式 + TaskQueue + DirWatcher
internal/handler/merge.go              业务逻辑移出 + typed response
internal/handler/config.go             业务逻辑移出 + typed response
internal/handler/status.go             typed response
internal/handler/auth.go               typed response
cmd/server/main.go                     瘦身 ~20 行
frontend/vite.config.ts                rollupOptions → rolldownOptions
frontend/tsconfig.json                 rootDir + paths
frontend/src/router/index.ts           guard 改写 + 认证状态迁移
frontend/src/views/settings/index.vue   1008 行 → ~80 行
frontend/src/layout/index.vue          拆分为壳 + 子组件
frontend/src/utils/http.ts             类型加固
```

### 删除文件

```
internal/utils/video.go            → ffprobe 迁入 ffmpeg 包
internal/utils/disk.go             → 迁入 fsutil
internal/utils/disk_windows.go     → 迁入 fsutil
```

---

## 七、预期效果

下面是完成后的代码库与现状对比：

| 指标 | 现在 | 重构后 |
|------|------|--------|
| 目录扫描重复 | 7 处 | 1 处（`fsutil.ScanStreamerDirs`） |
| 原子写入重复 | 4 处 | 1 处（`fsutil.AtomicWriteFile`） |
| 最大函数行数 | 172 行（`MergeService.Run`） | <50 行 |
| 最大组件行数 | 1008 行（Settings） | ~180 行（StorageTab） |
| CSS 类重复定义 | 6+ 处 | 0（全局 `_components.scss`） |
| 后端测试覆盖 | 仅叶子工具函数 | 核心 service 逻辑覆盖 |
| 前端测试覆盖 | 零 | composables + 工具函数 |
| 依赖版本差距 | Gin 落后 3 版 / Pinia 跨大版本 | 全部最新 |
| 并发控制 | 无（20 个 ffmpeg 同时跑） | 可配置并发上限 |
| 损坏文件处理 | 合并后才发现失败 | 合并前跳过 |
| 调度方式 | 固定 6 小时间隔 | 事件驱动 + 兜底 |
| 代码质量工具 | 无 | ESLint + golangci-lint + Vitest |
