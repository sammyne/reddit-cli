# AGENTS.md

## 项目概览

- **编程语言：** Go 1.24
- **项目名称：** rdt-cli

## Go 编码规范

- 使用 `go fmt` / `goimports` 格式化代码，不要手动调整格式。
- 遵循 [Effective Go](https://go.dev/doc/effective_go) 和 [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)。
- 导出的名称必须添加文档注释。未导出的名称在意图不明显时也应添加注释。
- 错误处理：始终显式检查错误，除非有充分的理由，否则不要使用 `_` 丢弃错误。
- 优先返回 `error` 而非 `panic`。仅在真正不可恢复的情况下使用 `panic`。
- 对于执行 I/O 或可能被取消的函数，使用 `context.Context` 作为第一个参数。
- 保持函数短小且职责单一。如果函数超过约 50 行，考虑拆分。

## 项目结构

```
cmd/           # CLI 入口
internal/      # 私有应用代码（不可被外部模块导入）
pkg/           # 公共库代码（如有）
```

- 将 `main` 包放在 `cmd/<二进制名称>/` 下。
- 使用 `internal/` 存放不应被外部导入的代码。
- 避免深层嵌套，优先使用扁平的包结构。

## 依赖管理

- 使用 Go modules（`go.mod` / `go.sum`）。添加或移除依赖后运行 `go mod tidy`。
- 优先使用标准库。仅在第三方依赖能提供显著价值时才引入。

## 测试

- 测试文件与被测代码放在同一目录：`foo.go` -> `foo_test.go`。
- 适当使用表驱动测试。
- 运行测试：`go test ./...`
- 运行 lint：`go vet ./...`

## 构建与运行

- 构建：`go build ./...`
- 运行测试：`go test ./...`
- Lint：`go vet ./...`

## 代码变更准则

- 遵循现有代码风格，不要格式化或重构无关代码。
- 每一行变更都应直接关联到当前任务。
- 移除因你的变更而变得未使用的 import/变量/函数；不要移除已有的未使用代码，除非被明确要求。
- 添加新 CLI 命令时，遵循现有的命令注册模式。
