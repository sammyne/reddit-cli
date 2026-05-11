# TODO

## 任务列表

### ✅ 1. 新建 `cmd/version.go`，实现版本信息获取逻辑

- 优先级: P0
- 依赖项: 无
- 涉及文件: `cmd/version.go`（新建）
- 验收标准:
  - 定义 `var version = "dev"`，可通过 ldflags 注入
  - 实现 `gitCommit()` 函数，通过 `runtime/debug.ReadBuildInfo()` 读取 `vcs.revision` 并截取前 7 位，不可用时返回 `"unknown"`
  - 实现 `fullVersion()` 函数，返回 `"{version}, build {short_hash}"` 格式的字符串
  - `go vet ./...` 通过
- 风险/注意点: `runtime/debug.ReadBuildInfo()` 在非 VCS 环境下构建时 `vcs.revision` 可能为空，需正确回退

### ✅ 2. 更新 `cmd/root.go`，使用新的版本逻辑

- 优先级: P0
- 依赖项: 1
- 涉及文件: `cmd/root.go`
- 验收标准:
  - 移除 `const version = "0.4.1"`
  - 在 `init()` 中设置 `rootCmd.Version = fullVersion()`（因为 `fullVersion()` 依赖 `runtime/debug`，不能在包级别 var 初始化时使用，需在 `init` 中赋值）
  - `go build ./...` 和 `go vet ./...` 通过
- 风险/注意点: `rootCmd` 的 `Version` 字段需要从包级别初始化改为在 `init()` 中设置

### ✅ 3. 更新 `.github/workflows/release.yml`，通过 ldflags 注入 version

- 优先级: P0
- 依赖项: 1, 2
- 涉及文件: `.github/workflows/release.yml`
- 验收标准:
  - build step 的 `go build` 命令增加 `-ldflags "-X github.com/sammyne/reddit-cli/cmd.version=${VERSION}"`，其中 `VERSION` 为去掉 `v` 前缀的 tag 值
  - YAML 语法正确
- 风险/注意点: `actions/checkout@v4` 的 `fetch-tags` 默认行为在 tag 触发的 workflow 中是可用的，无需额外配置

## 实现建议

- `cmd/version.go` 保持精简，仅包含版本相关逻辑
- `gitCommit()` 遍历 `BuildInfo.Settings` 查找 `vcs.revision` key，截取前 7 位即可
- CI 中已有 `VERSION="${GITHUB_REF_NAME#v}"`，可直接复用该变量传给 ldflags
