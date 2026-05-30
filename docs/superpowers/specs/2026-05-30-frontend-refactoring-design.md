# Frontend Refactoring Design

## 2026-05-30

## Overview

重构 Bililive Helper 前端：从单文件 Vue 3 CDN SPA 迁移到 Vite + Vue 3 完整组件化架构。

**目标：**
- 组件化拆分（8+ 个 Vue 组件）
- Vite 5 构建工具链
- TypeScript（可选，后续引入）
- Composable 模式复用逻辑
- 保持现有的 UI 设计和 CSS 变量系统

**约束：**
- 不改变功能行为
- 保持暗黑主题支持
- Go 后端需要适配静态文件服务
- 迁移期间保持应用可运行

---

## Architecture

### 技术栈

| Layer | Technology |
|-------|-----------|
| Build | Vite 5 |
| Framework | Vue 3 (Composition API + `<script setup>`) |
| Language | JavaScript (TypeScript optional) |
| Styling | CSS Variables (existing design system) |
| Deployment | Vite build → Go static server |

### 项目结构

```
/frontend
├── index.html
├── vite.config.js
├── package.json
└── src/
    ├── main.js
    ├── App.vue
    ├── components/
    │   ├── TopNav.vue
    │   ├── StatusDashboard.vue
    │   ├── StreamerList.vue
    │   ├── TaskHistory.vue
    │   ├── LogViewer.vue
    │   ├── ConfigPanel.vue
    │   ├── TaskScheduler.vue
    │   └── ManualMerge.vue
    ├── composables/
    │   ├── useApi.js
    │   ├── useStreamerData.js
    │   ├── useConfig.js
    │   ├── useTaskRunner.js
    │   └── useToast.js
    └── styles/
        ├── variables.css
        └── base.css
```

### 构建流程

**开发：**
```bash
cd frontend && npm run dev
# Vite dev server on localhost:5173
# Proxy /api/* to Go backend on :5000
```

**生产：**
```bash
cd frontend && npm run build
# Output to /dist
# Copy to /frontend/ in Go project root
```

**Go 后端适配：**
- 服务静态文件从 `/frontend/`
- 开发时通过 Vite proxy 到 Go API

---

## Components

### 1. TopNav.vue (~30 lines)

**职责：** 顶部导航栏

**Props:** None
**Events:** None
**Dependencies:** `useTaskRunner`

**功能：**
- Logo + 标题（"Bililive Helper"）
- 运行状态指示器（当 running=true 时显示）
- 退出登录链接

---

### 2. StatusDashboard.vue (~120 lines)

**职责：** 系统状态仪表盘

**Props:** None
**Events:** `run(task, streamer)` - 触发任务执行
**Dependencies:** `useStreamerData`, `useTaskRunner`

**功能：**
- 磁盘占用量 gauge（百分比 + 进度条）
- 自动合并/清理状态指示器
- 下次执行时间显示
- 全局合并/清理按钮

**子组件：**
- 无（自包含）

---

### 3. StreamerList.vue (~180 lines)

**职责：** 主播列表展示和管理

**Props:** None
**Events:** `run(task, streamer)`, `openManualMerge(streamer)`
**Dependencies:** `useStreamerData`, `useTaskRunner`

**功能：**
- 搜索框（实时过滤）
- 可排序表格（主播名、占用大小、文件数）
- 分页逻辑（每页 10 条）
- 操作按钮（合并/清理/手动合并）
- 空状态显示

**内部状态：**
- 当前页码
- 搜索关键词
- 排序字段和方向

---

### 4. TaskHistory.vue (~80 lines)

**职责：** 操作历史记录

**Props:** None
**Events:** None
**Dependencies:** `useApi`

**功能：**
- 任务类型过滤（合并/清理/全部）
- 分页历史表格（时间、详情、状态）
- 状态图标显示（✓ 成功，✗ 失败）

**内部状态：**
- 当前页码
- 过滤任务类型

---

### 5. LogViewer.vue (~100 lines)

**职责：** 系统日志查看器

**Props:** None
**Events:** None
**Dependencies:** `useApi`

**功能：**
- 日志类型选择（合并/清理）
- 文件选择下拉框
- 实时日志显示（语法高亮）
- 自动滚动开关
- 终端风格 UI（暗黑背景）

**内部状态：**
- 当前日志类型
- 当前日志文件
- 日志内容

---

### 6. ConfigPanel.vue (~250 lines)

**职责：** 配置管理表单

**Props:** None
**Events:** `save()`, `reset()`
**Dependencies:** `useConfig`, `TaskScheduler`

**功能：**
- 两大部分：
  1. 基础策略（存储目录、白名单、阈值等）
  2. 任务调度（自动合并/清理开关、间隔、静默时段）
- 输入验证（数值范围、必填项）
- 保存/重置按钮
- 帮助说明折叠

**子组件：**
- TaskScheduler.vue（约 80 行，处理调度配置）

**内部状态：**
- 配置对象
- 表单验证状态

---

### 7. TaskScheduler.vue (~80 lines)

**职责：** 任务调度配置

**Props:** `config` (from ConfigPanel)
**Events:** None
**Dependencies:** 无

**功能：**
- 自动合并开关 + 间隔时间
- 自动清理开关 + 间隔时间
- 静默时段配置（开始时间 - 结束时间）
- 开关控制禁用状态

**内部状态：**
- 无（纯展示和输入）

---

### 8. ManualMerge.vue (~100 lines)

**职责：** 手动合并弹窗

**Props:** `visible`, `streamer`
**Events:** `close()`, `merge(files)`
**Dependencies:** `useApi`, `useTaskRunner`

**功能：**
- 文件列表表格（文件名、大小、类型）
- 全选/反选复选框
- 已选统计（数量 + 总大小）
- 清空/开始合并按钮
- 加载/空状态显示

**内部状态：**
- 文件列表
- 已选文件

---

## Composables

### useApi.js (~60 lines)

**职责：** API 调用封装

**返回：**
```javascript
{
  get(url): Promise<any>
  post(url, data): Promise<any>
  stream(url, options): Promise<ReadableStream>  // SSE 支持
}
```

**特性：**
- 统一错误处理
- 401 自动重定向到 /login
- Content-Type 自动识别
- JSON 解析

---

### useStreamerData.js (~80 lines)

**职责：** 主播数据管理

**返回：**
```javascript
{
  streamers: Ref<Array>
  diskUsage: Ref<number>
  totalGB: Ref<number>
  detail: Ref<Object>
  
  fetchStatus(): Promise<void>
  fetchDetail(): Promise<void>
  filtered(keyword): Array
  sorted(by, asc): Array
}
```

**特性：**
- 实时数据获取
- 缓存和更新策略
- 计算属性（过滤、排序）

---

### useConfig.js (~70 lines)

**职责：** 配置管理

**返回：**
```javascript
{
  config: Ref<Object>
  schedule: Ref<Object>
  
  fetchConfig(): Promise<void>
  saveConfig(): Promise<void>
  saveSchedule(): Promise<void>
  resetSchedule(): Promise<void>
}
```

**特性：**
- 配置验证
- 自动保存
- 默认值处理

---

### useTaskRunner.js (~90 lines)

**职责：** 任务执行和进度跟踪

**返回：**
```javascript
{
  running: Ref<boolean>
  progress: Ref<number>
  progressLabel: Ref<string>
  logs: Ref<string>
  
  run(task, streamer): void           // SSE 实时流
  startManualMerge(files): Promise<void>
  stop(): void
}
```

**特性：**
- SSE 事件源连接
- 实时进度更新
- 日志流处理
- 任务取消支持

---

### useToast.js (~30 lines)

**职责：** 消息提示

**返回：**
```javascript
{
  toast(message, type): void  // type: 'ok' | 'err' | 'info'
}
```

**特性：**
- 自动消失（3.5s）
- 多消息队列
- DOM 操作管理

---

## Data Flow

### App.vue (Root)

```
provide({
  running: useTaskRunner().running,
  progress: useTaskRunner().progress,
  progressLabel: useTaskRunner().progressLabel,
  logs: useTaskRunner().logs,
  toast: useToast().toast
})

// 子组件 inject 使用
```

### 事件传递

```
StreamerList.vue
  └─ @run(task, streamer)
     └─ App.vue (接收)
        └─ useTaskRunner.run(task, streamer)

ConfigPanel.vue
  └─ @save()
     └─ App.vue (接收)
        └─ useConfig.saveConfig()
```

### Props 向下

```
ManualMerge.vue
  └─ :visible="showMergeModal"
  └─ :streamer="selectedStreamer"
  └─ @close="showMergeModal = false"
  └─ @merge="handleMerge"
```

---

## Error Handling

### API 层

```javascript
// useApi.js
async function get(url) {
  try {
    const response = await fetch(url, options);
    if (response.status === 401) {
      window.location.href = '/login';
      throw new Error('未登录');
    }
    return await response.json();
  } catch (error) {
    toast(error.message, 'err');
    throw error;
  }
}
```

### SSE 流

```javascript
// useTaskRunner.js
const es = new EventSource(url);
es.onerror = () => {
  running.value = false;
  logs.value += '\n❌ 任务执行中断\n';
  es.close();
};
```

### 表单验证

```javascript
// ConfigPanel.vue
function validateConfig(config) {
  if (config.TRIGGER_THRESHOLD < 0 || config.TRIGGER_THRESHOLD > 100) {
    toast('阈值必须在 0-100 之间', 'err');
    return false;
  }
  // ... 其他验证
  return true;
}
```

---

## Performance Optimizations

### 代码拆分

```javascript
// vite.config.js
export default defineConfig({
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          vue: ['vue'],
          // 按需拆分其他库
        }
      }
    }
  }
})
```

### 懒加载

```javascript
// App.vue
const ConfigPanel = defineAsyncComponent(() =>
  import('./components/ConfigPanel.vue')
)
```

### 防抖

```javascript
// StreamerList.vue
import { refDebounced } from '@vueuse/core';

const searchQuery = ref('');
const debouncedQuery = refDebounced(searchQuery, 300);
```

---

## Migration Steps

### Phase 1: Setup Vite Project
1. 创建 `/frontend` 目录
2. 初始化 Vite + Vue 3 项目
3. 迁移 CSS 变量和全局样式
4. 迁移 Vue 3 CDN 到 Vite

### Phase 2: Extract Composables
1. 抽取 API 调用逻辑 → useApi
2. 抽取数据管理逻辑 → useStreamerData
3. 抽取配置管理逻辑 → useConfig
4. 抽取任务执行逻辑 → useTaskRunner

### Phase 3: Create Components
1. 创建基础组件（TopNav）
2. 创建主要组件（StatusDashboard, StreamerList）
3. 创建辅助组件（TaskHistory, LogViewer）
4. 创建配置组件（ConfigPanel, TaskScheduler）
5. 创建弹窗组件（ManualMerge）

### Phase 4: Integrate and Test
1. 在 App.vue 集成所有组件
2. 功能测试
3. 响应式测试
4. 暗黑主题测试

### Phase 5: Go Backend Adaptation
1. 修改 Go 静态文件服务
2. 适配生产构建路径
3. 部署测试

---

## Testing Strategy

### Unit Tests (Optional)
- Composables 单元测试
- 工具函数测试

### Integration Tests
- 组件间数据流测试
- API 交互测试
- SSE 流测试

### E2E Tests
- 登录流程
- 主播列表操作
- 配置保存
- 任务执行

---

## Rollback Plan

如果迁移失败：
1. 保留旧的 `/frontend` 目录（重命名为 `/frontend-legacy`）
2. 恢复 Go 后端配置
3. 回滚时间：~30 分钟

---

## Success Criteria

1. ✅ 功能完整：所有现有功能都正常工作
2. ✅ UI 一致：视觉效果和交互完全一致
3. ✅ 代码清晰：组件拆分合理，可维护性提升
4. ✅ 构建正常：Vite dev 和 build 都正常
5. ✅ 性能无损：加载时间不增加
6. ✅ 暗黑主题：完美支持

---

## References

- Vue 3 Composition API: https://vuejs.org/guide/extras/composition-api-faq.html
- Vite 5: https://vitejs.dev/
- Vue Use: https://vueuse.org/
- 现有代码: `/templates/index.html` (934 行)
