# bililive-helper-go 深度重构设计文档

> 日期：2026-06-14 | 方案：C — 深度重构 | 外部行为不变

## 目标

1. 消除跨文件重复代码（目录扫描 7 处、原子写入 4 处、CSS 类 6+ 处）
2. 核心业务逻辑可测试（通过接口抽象实现 mock）
3. 理清职责边界（业务逻辑归 service、handler 只做 HTTP 适配）
4. 拆分超大函数和 God Component
5. 前后端补全工程化基础设施（ESLint、Vitest、Go 单元测试）
6. 依赖升级到最新版本（Go 1.26、Gin 1.12、Pinia 3、Vue Router 5、Vite 8、TypeScript 6）
7. 任务队列与并发控制（可配置并发上限，避免 I/O 争抢）
8. 文件健康预检（合并前 ffprobe 快检，跳过损坏文件）
9. 事件驱动调度（目录变化检测替代纯固定间隔，文件更快被处理）

## 约束

- 外部 API 端点路径、参数格式不变
- config.json / .credentials.json / schedule.json 格式不变
- Docker 部署方式和环境变量名不变
- 数据库不在本次范围内（保留文件存储）

## 技术栈评估结论

当前技术栈（Go + Gin + Vue 3 + Element Plus + Vite）是本项目的最佳选择，不需要更换。
重构的投入放在"如何写"而不是"用什么写"上。详见技术栈评估分析。

## 方法论

本次重构遵循以下行业最佳实践：

| 方法论 | 来源 | 在本项目中的应用 |
|--------|------|-----------------|
| **特征测试先行** | Michael Feathers《Working Effectively with Legacy Code》 | 在任何重构之前，为关键路径写 Characterization Tests，记录当前行为作为基线 |
| **绞杀者模式** | Martin Fowler / Paul Hammant | 新代码与旧代码短暂共存（如 fsutil 替代 utils 中的函数），逐个迁移验证后再删除旧代码 |
| **Mikado 依赖图** | Ola Ellnestam & Daniel Brolund | 对每个大阶段画出内部依赖关系，优先完成叶子节点（无依赖的原子改动） |
| **先让改动变简单** | Kent Beck "Make the change easy, then make the easy change" | 每个阶段先问"这个改动要变得简单需要什么条件"，先创造条件 |
| **ADR 决策记录** | Michael Nygard | 每个重要架构决策记录为 ADR，存放在 `docs/adr/` |
| **风险量化** | Fowler 技术债务象限 | 用 golangci-lint + gocyclo 建立静态分析基线，用数据驱动优先级 |
| **API 契约验证** | Golden Master / Contract Testing | 重构前后端前，先记录 API 响应基线，重构后对比验证 |
| **小步提交** | Fowler "Refactoring" | 每个叶子节点改动独立提交，代码库始终可编译可运行 |

### 安全网层次

```
Level 0 — 静态分析基线（golangci-lint + gocyclo + vue-tsc）
Level 1 — 特征测试（记录当前行为，重构后对比）
Level 2 — API 响应基线（Golden Master，前后端兼容性保障）
Level 3 — 单元测试（重构过程中逐步补充）
Level 4 — 手动冒烟测试（每个阶段完成后）
```

### ADR 格式

每个重要决策记录为 `docs/adr/NNNN-<title>.md`：
```markdown
# ADR-0001: 提取 fsutil 包替代散落的文件操作

## Status: Accepted
## Date: 2026-06-14
## Context: 原子写入模式在 4 个文件中重复，目录扫描在 7+ 处独立实现
## Decision: 新建 internal/fsutil 包统一文件操作，逐个迁移调用点
## Consequences: 正面 — 消除重复、统一行为；负面 — 短暂的新旧共存期
```

---

## 第负一阶段：建立安全网（重构前必做）

> Michael Feathers: "Legacy code is code without tests. You cannot safely refactor code you cannot test."
> 在任何代码改动之前，先建立行为基线。这不是"补测试"，而是"记录当前行为作为安全网"。

### -1.1 静态分析基线

```bash
# Go 静态分析
golangci-lint run ./...               # 代码质量基线
gocyclo -over 15 ./internal/...       # 圈复杂度（当前 MergeService.Run 等超大函数会标红）

# 前端类型检查
cd frontend && npx vue-tsc --noEmit   # 类型错误基线
```

记录输出作为"before"快照，重构后对比确保不引入新问题。

### -1.2 后端特征测试（Characterization Tests）

为以下核心路径写特征测试 — 不验证"正确"，只记录"当前行为"：

| 目标 | 测试内容 | 方法 |
|------|---------|------|
| `MergeService.scanTasks` | 扫描一组模拟目录，记录返回的分类结果 | 创建临时目录结构，调用函数，断言输出 |
| `MergeService.SortByFilename` | 已有测试，扩展边界用例 | 补充空列表、单文件、跨天分片等 case |
| `CleanService.collectCandidates` | 给定一组文件和配置，记录清理优先级 | 模拟不同大小/年龄的文件 |
| `config.Validate` | 给定边界值，记录验证结果 | 包括有效和无效输入 |
| `config.Apply` | 给定修改，记录原子写入行为 | 验证文件确实被原子更新 |
| `ParseFilename` | 已有测试，确认覆盖率 | 确认 bililive-go 各种命名格式都被覆盖 |

### -1.3 API 响应基线（Golden Master）

为所有 API 端点记录当前响应格式作为基线：

```bash
# 启动服务，用 curl 记录所有端点的响应
curl -s http://localhost:5000/api/health > tests/baseline/api_health.json
curl -s http://localhost:5000/api/status > tests/baseline/api_status.json
# ... 对所有端点重复
```

重构后运行相同命令，diff 对比确保 API 契约不变。

### -1.4 前端快照基线

```bash
cd frontend
npx vue-tsc --noEmit > ../tests/baseline/vue-tsc-before.txt 2>&1
npm run build 2>&1 | tee ../tests/baseline/build-before.txt
```

---

## 第零阶段：依赖升级

> 原因：先升后改 — 如果先重构代码再升级依赖，重构时基于旧 API 写的代码在新版本中可能又需要改。
> Pinia 3 + Vue Router 5 是 breaking change，前端重构（阶段六）的 composables 和 store 设计应直接基于新 API。

### 0.1 后端依赖升级

| 依赖 | 当前 | 目标 | Breaking Changes | 代码改动 |
|------|------|------|-----------------|---------|
| Go | 1.25 | 1.26 | 无 | go.mod 更新版本号 |
| Gin | v1.9.1 | v1.12.0 | 无（`interface{}` → `any`，不影响本项目） | 无 |
| Zap | v1.26.0 | v1.28.0 | 无（纯增量：`WithLazy`、`CheckPreWriteHook`） | 无 |
| x/crypto | v0.52.0 | v0.53.0 | 无（bcrypt API 稳定） | 无 |
| sessions | v0.0.5 | v1.1.0 | 跨大版本，但 cookie store API（`NewStore`/`Sessions`/`Default`/`Get`/`Set`/`Clear`/`Save`）不变 | 无代码改动，需测试 session cookie 兼容性（升级后让用户重新登录即可） |

执行：
```bash
go get github.com/gin-gonic/gin@v1.12.0 go.uber.org/zap@v1.28.0 golang.org/x/crypto@latest github.com/gin-contrib/sessions@v1.1.0
go mod tidy
go build ./... && go test ./...
```

### 0.2 前端依赖升级（分步，每步验证）

**第 1 步 — 无 breaking change 的 patch/minor 更新**

| 依赖 | 当前 | 目标 | 说明 |
|------|------|------|------|
| Vue | ^3.5.0 | ^3.5.0 | semver 兼容，`npm update` 即可 |
| Axios | ^1.7.0 | ^1.7.0 | semver 兼容 |
| Element Plus | ^2.9.0 | ^2.14.2 | 次版本更新。**唯一注意点**：`el-teleport` 组件在 2.14.0 被移除（本项目未使用，安全）；disabled 优先级变化（v2.12.0，需检查表单禁用逻辑） |
| Sass | ^1.80.0 | ^1.80.0 | 已满足 Vite 8 要求的 >=1.70.0 |

```bash
npm update vue axios element-plus sass
npm run build   # 验证
```

**第 2 步 — Vite 6 → 8（构建工具）**

Breaking Changes：
- `build.rollupOptions` → `build.rolldownOptions`（Rolldown 替代 Rollup/esbuild）
- JS 压缩从 esbuild 换为 Oxc，CSS 压缩换为 Lightning CSS（无自定义配置则自动适配）
- 需要 Node.js 20.19+ 或 22.12+

需改动 `vite.config.ts`：
```diff
  build: {
    outDir: "../templates",
    emptyOutDir: false,
-   rollupOptions: {
+   rolldownOptions: {
      output: {
        chunkFileNames: "assets/[name]-[hash].js",
        entryFileNames: "assets/[name]-[hash].js",
        assetFileNames: "assets/[name]-[hash].[ext]"
      }
    }
  }
```

```bash
npm install vite@^8.0.0 @vitejs/plugin-vue@^6.0.0
# 验证 unplugin-auto-import 和 unplugin-vue-components 兼容 Vite 8
npm run build   # 验证
```

**第 3 步 — TypeScript 5 → 6**

Breaking Changes：
- `baseUrl` 不再作为模块查找根 → `paths` 需要改为相对路径
- `rootDir` 默认值变化 → 需显式指定
- `types` 默认改为 `[]`（空数组）→ 本项目已显式设为 `["vite/client"]`，无影响

需改动 `tsconfig.json`：
```diff
  {
    "compilerOptions": {
      "target": "ES2020",
      "module": "ESNext",
      "moduleResolution": "bundler",
      "strict": true,
+     "rootDir": "./src",
      "jsx": "preserve",
      "resolveJsonModule": true,
      "isolatedModules": true,
      "esModuleInterop": true,
      "lib": ["ES2020", "DOM", "DOM.Iterable"],
      "skipLibCheck": true,
      "noEmit": true,
      "baseUrl": ".",
      "paths": {
-       "@/*": ["src/*"]
+       "@/*": ["./src/*"]
      },
      "types": ["vite/client"]
    }
  }
```

```bash
npm install typescript@^6.0.0
npx vue-tsc --noEmit   # 验证类型检查通过，修复可能的新错误
```

**第 4 步 — Pinia 2 → 3 + Vue Router 4 → 5**

Pinia 3 Breaking Changes：
- 移除 `defineStore({ id })` 旧语法（本项目未使用，安全）
- 移除 `PiniaStorePlugin` 类型（本项目未使用，安全）
- 要求 Vue 3.5+ 和 TypeScript 5+（已满足）
- 本项目的 setup store 模式（`defineStore('id', () => {...})`）**完全不受影响**

Vue Router 5 Breaking Changes：
- `next()` 回调标记 deprecated（仍可用，建议改为 return-value 模式）
- 本项目不使用 `NavigationResult`，无影响

需改动 `router/index.ts`：
```diff
- router.beforeEach(async (to, _from, next) => {
+ router.beforeEach(async (to, _from) => {
    document.title = ...
-   if (to.meta.public) { next(); return; }
+   if (to.meta.public) return true;
    // ...
-   next("/login");
+   return "/login";
  });
```

```bash
npm install pinia@^3.0.0 vue-router@^5.0.0
npm run build   # 验证
```

### 0.3 前端依赖升级风险总结

| 风险 | 依赖 | 说明 |
|------|------|------|
| 🟢 零风险 | Zap, x/crypto, Vue, Axios, Sass | 纯增量更新 |
| 🟢 极低 | Gin, Pinia 3, Element Plus 2.14 | API 不受影响 |
| 🟡 低 | sessions v1.1.0 | API 兼容，测试 cookie 兼容 |
| 🟡 低 | Vue Router 5 | 改 guard 写法，~5 分钟 |
| 🟡 中 | TypeScript 6 | tsconfig 改动 + 可能触发新类型错误 |
| 🟡 中 | Vite 8 | rollupOptions → rolldownOptions + 验证 unplugin 兼容 |

---

## 第一阶段：后端基础设施层

> 应用 Strangler Fig 模式：新代码与旧代码短暂共存，逐个迁移调用点，验证后再删除旧代码。

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
- **文件健康预检**：扫描时同步调用 `ffprobe` 快检（文件大小 > 0 且可读取时长），返回 `FileHealth` 标记（Healthy / Corrupt / Skippable），合并时直接跳过损坏文件，避免浪费几十分钟的 ffmpeg 处理
- **迁移策略（绞杀者模式）**：先建 fsutil，在一处调用点替换（如 scheduler.go），运行特征测试验证，确认无误后再替换下一处。旧函数暂时保留但标记 `// DEPRECATED: use fsutil.XXX`
- 所有调用点迁移完成后，删除旧代码和 utils 中的冗余文件

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

> 应用 Mikado 方法：此阶段内部依赖复杂，先做叶子节点，逐步向上。

### Mikado 依赖图

```
目标：MergeService/CleanService 可测试 + 函数 <50 行 + 任务队列 + 事件驱动
│
├── 叶子节点（无依赖，先做）：
│   ├── 2.4a collectCandidates 返回值重构（指针切片 → 返回值）
│   ├── 2.4b deleteFiles 拆分为 snapshot + filter + delete
│   └── 2.2a 从 utils/video.go 迁移 ffprobe 函数到 ffmpeg 包
│
├── 中间节点（依赖叶子）：
│   ├── 2.3 ffmpeg.Executor 接口定义 + DefaultExecutor 实现
│   ├── 2.1 TaskRunner 接口定义
│   ├── 2.1a TaskQueue 队列实现（依赖 TaskRunner）
│   ├── 2.1b DirWatcher 目录监听实现
│   └── 2.2b MergeService.Run() 拆分（需要 2.3 的接口）
│
└── 根节点（依赖中间）：
    ├── 2.5 业务逻辑移出 Handler（需要 2.1 + 2.3 的接口）
    ├── MergeService/CleanService 实现 TaskRunner 接口
    └── SchedulerService 改造：TaskQueue + DirWatcher + 兜底定时器
```

每个叶子节点改动独立提交后运行特征测试验证。

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

### 2.1a 任务队列与并发控制

当前状态：20 个主播同时下播会同时触发 20 个合并进程，I/O 争抢导致全部变慢。

新增设计：
```go
type TaskQueue struct {
    sem     chan struct{}          // 并发信号量，如 make(chan struct{}, 3)
    queue   chan TaskRequest       // 等待队列
    running sync.Map              // 当前运行中的任务
}

type TaskRequest struct {
    Runner   TaskRunner
    Priority int                   // 可选优先级
    EnqueuedAt time.Time
}

func (q *TaskQueue) Submit(req TaskRequest)  // 提交到队列
func (q *TaskQueue) Run(ctx context.Context)  // 消费循环
```

- 用户可配置并发上限（如 `max_concurrent_tasks: 3`）
- 超出并发数的任务排队等待
- 配置项加入 config.json（兼容现有格式）

### 2.1b 事件驱动调度（替代纯固定间隔）

当前状态：定时器每 6 小时触发一次，主播 6 点下播要等到凌晨 12 点才合并。

新增设计：
```go
type SchedulerService struct {
    // ... 现有字段 ...
    watcher    *DirWatcher        // 目录变化监听
    debounce   time.Duration      // 防抖时间（默认 5 分钟）
}

type DirWatcher struct {
    root     string
    interval time.Duration        // 轮询间隔（轻量实现，不用 fsnotify）
    onChange func(streamer string) // 检测到变化时的回调
}
```

- 实现方式：轻量轮询（每分钟检查各主播目录的修改时间），比 fsnotify 更可靠（Docker/网络文件系统下 fsnotify 不稳定）
- 检测到"某主播目录 N 分钟无变化" → 认为录制结束 → 提交合并任务到队列
- 固定间隔降级为"兜底"（每天凌晨跑一次确保无遗漏）
- 防抖时间可配置（防止录制中途短暂暂停被误判为结束）

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
阶段 -1 — 建立安全网                            ← 最最优先，任何代码改动之前
  -1.1 静态分析基线（golangci-lint + gocyclo + vue-tsc）
  -1.2 后端特征测试（核心路径的 Characterization Tests）
  -1.3 API 响应基线（Golden Master）
  -1.4 前端构建基线

阶段 0 — 依赖升级                                ← 基于最新 API
  0.1 后端：Go 1.26 + Gin 1.12 + Zap 1.28 + sessions 1.1
  0.2 前端分步：Element Plus 2.14 → Vite 8 → TypeScript 6 → Pinia 3 + Vue Router 5

阶段 1 — fsutil + taskctx（基础设施）             ← 绞杀者模式，逐点迁移
阶段 2 — service 接口化 + 函数拆分                ← Mikado 依赖图驱动，叶子节点优先
阶段 3 — config 系统重构                          ← 独立模块，影响面可控
阶段 4 — handler 层重构 + 路由拆分                ← 依赖阶段 2 的接口
阶段 5 — main.go 拆分                             ← 依赖阶段 4 的路由注册
阶段 6 — 前端组件拆分 + composables               ← 前后端可并行，基于 Pinia 3 / Vue Router 5 API
阶段 7 — 前端工程化（ESLint/Vitest/CSS）          ← 依赖阶段 6，基于 Vite 8
阶段 8 — 后端测试补全 + 静态分析对比              ← 依赖阶段 2-3 的接口，对比阶段 -1 基线
```

**关键规则：**
- 每个叶子节点改动独立提交，代码库始终可编译可运行
- 每阶段完成后运行特征测试（阶段 -1 建立的），确认行为未变
- 阶段 8 完成后对比 golangci-lint / gocyclo 输出，确认复杂度下降
- 重构提交和功能提交严格分开（不在同一个 commit 里既重构又改功能）

---

## 文件变更概览

### 阶段 -1 安全网涉及文件
```
tests/baseline/                     — 新建目录，存放所有基线文件
tests/baseline/api_*.json           — API 响应基线（Golden Master）
tests/baseline/vue-tsc-before.txt   — 前端类型检查基线
tests/baseline/build-before.txt     — 前端构建基线
tests/baseline/lint-before.txt      — Go 静态分析基线
internal/service/*_test.go          — 特征测试（扩展已有 + 新建）
docs/adr/                           — 架构决策记录目录
docs/adr/0001-fsutil-package.md     — 首个 ADR
```

### 阶段 0 升级涉及文件
```
go.mod                              — Go 版本 + 4 个依赖版本升级
go.sum                              — 自动生成
frontend/package.json               — 6 个依赖版本升级
frontend/package-lock.json          — 自动生成
frontend/vite.config.ts             — rollupOptions → rolldownOptions
frontend/tsconfig.json              — rootDir + paths 调整
frontend/src/router/index.ts        — guard next() → return（与阶段 6.4 合并）
```

### 新增文件
```
internal/fsutil/atomic.go
internal/fsutil/scan.go              — 含文件健康预检（ffprobe 快检）
internal/fsutil/safeio.go
internal/fsutil/path.go
internal/taskctx/taskctx.go
internal/service/interfaces.go       — TaskRunner 接口
internal/service/taskqueue.go        — 任务队列与并发控制
internal/service/dirwatcher.go       — 目录变化监听（事件驱动调度）
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
internal/service/scheduler.go      — 注册模式替代 switch + TaskQueue + DirWatcher 事件驱动
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
