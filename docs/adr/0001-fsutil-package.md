# ADR-0001: 提取 fsutil 包统一文件操作

## 状态

Accepted

## 日期

2026-06-14

## 背景

项目中存在多处重复的文件操作代码：
- **原子写入**（write tmp → fsync → rename）在 config.go、scheduler.go、history.go 中重复 4 次
- **目录扫描**（遍历 TARGET_DIR 下的主播子目录）在 handler/merge.go、handler/status.go、handler/config.go、service/merge.go、service/clean.go、service/scanner.go 等 7+ 处独立实现
- **安全删除**（带重试）和**路径校验**（防遍历攻击）散落在 utils 包中

这种重复导致：
1. 改一次目录结构需求要改 7 处，极易遗漏
2. 各处实现细节不一致（有的有 fsync 有的没有）
3. 无法统一添加新功能（如文件健康预检）

## 决策

新建 `internal/fsutil` 包，包含：
- `atomic.go` — `AtomicWriteFile(path, data, perm) error`
- `scan.go` — `ScanStreamerDirs(root) ([]StreamerDir, error)` + 文件健康预检
- `safeio.go` — `SafeUnlink(path) error`
- `path.go` — `ValidatePath(root, target) error`

采用绞杀者模式迁移：新旧代码短暂共存，逐个迁移调用点，验证后再删除旧代码。

## 后果

**正面：**
- 消除 4 处原子写入重复和 7+ 处目录扫描重复
- 统一行为（所有原子写入都包含 fsync）
- 为文件健康预检提供自然的集成点

**负面：**
- 短暂的新旧代码共存期（约 1-2 周）
- 需要逐个验证迁移点，工作量比"一次性替换"大
- 删除旧代码时需要确认没有遗漏的调用点
