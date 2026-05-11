# 需求

## 目标与背景

当前 `reddit-cli` 的 `--version` 选项输出硬编码版本号 `0.4.1`（定义在 `cmd/root.go` 中的常量 `version`）。需要将其改为动态输出：**当前所在分支的最新 git tag + 当前的 git 短版本哈希**，格式与 Docker 一致。

Docker 样例：`Docker version 26.1.3, build b72abbb`
本项目预期：`reddit version 0.4.1, build b72abbb`

这样做的好处是：
- 版本信息可准确追溯到具体的代码提交
- 不再需要每次发布时手动更新硬编码的版本号

## 功能需求列表

### 核心功能

- 将 `cmd/root.go` 中的 `version` 从 `const` 改为 `var`，默认值为 `"dev"`
- **build 字段（git commit hash）通过 `runtime/debug.ReadBuildInfo()` 自动获取**，读取 `vcs.revision` 并截取前 7 位作为短哈希，不依赖 ldflags
  - Go 在 `go build` 时会自动将 VCS 信息嵌入二进制文件（Go 1.18+，默认启用）
  - 若 `vcs.revision` 不可用，回退到 `"unknown"`
- `version`（tag）仍通过 `go build -ldflags -X` 注入，因为 `runtime/debug` 不提供 tag 信息
- cobra 的 `Version` 字段设为 `{version}, build {short_hash}`，配合 cobra 默认模板输出 `reddit version 0.4.1, build b72abbb`
- 更新 `.github/workflows/release.yml` 的构建命令，使用 `-ldflags` 注入 `version`（仅 tag）

### 扩展功能

- 无

## 非功能需求

- **性能**：无额外影响，版本信息在编译时注入
- **安全**：无安全影响
- **兼容性**：`go build .` 不带 ldflags 时，版本输出为 `reddit version dev, build b72abbb`（build 字段自动从 VCS 获取）；若不在 git 仓库中构建则为 `reddit version dev, build unknown`
- **可维护性**：移除硬编码版本号，减少发布时的手动操作
- **测试要求**：`go build ./...` 和 `go vet ./...` 通过

## 边界与不做事项

- 不引入 Makefile 或 goreleaser 等额外构建工具
- 不添加 `vcs.modified`（dirty 状态）检测
- 不修改 cobra 的版本输出模板（使用默认的 `{name} version {version}` 模板，通过组合 Version 字段内容实现目标格式）

## 假设与约束

- **技术假设**：`version`（tag）通过 `-ldflags -X` 注入；`build`（commit hash）通过 `runtime/debug.ReadBuildInfo()` 自动获取，不依赖 ldflags
- **格式假设**：版本输出格式与 Docker 一致，即 `reddit version {tag_without_v}, build {short_hash}`
- **环境约束**：CI 环境中 `actions/checkout@v4` 默认会拉取 tags，可通过 `fetch-tags: true` 确保

## 待确认事项

- 无
