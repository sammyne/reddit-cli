# TODO

## 任务列表

### ✅ 1. 创建 workflow 文件，配置 tag 触发和测试检查 job

- 优先级: P0
- 依赖项: 无
- 涉及文件: `.github/workflows/release.yml`（新建）
- 验收标准:
  - workflow 文件存在且 YAML 语法正确
  - `on.push.tags` 配置为 `v*` 格式
  - `test` job 在 `ubuntu-latest` 上运行，安装与 `go.mod` 一致的 Go 版本
  - `test` job 依次执行 `go vet ./...` 和 `go test ./...`
- 风险/注意点: Go 版本需与 `go.mod` 中的 `go 1.26.3` 保持一致，确认 `actions/setup-go` 支持该版本

### ✅ 2. 添加 build job，使用 matrix 在原生 runner 上编译并打包产物

- 优先级: P0
- 依赖项: 1
- 涉及文件: `.github/workflows/release.yml`
- 验收标准:
  - `build` job 依赖 `test` job（`needs: test`），测试失败时不执行构建
  - matrix 包含 3 个目标，每个目标指定对应的原生 runner：
    - `linux/amd64` → `ubuntu-latest`
    - `linux/arm64` → `ubuntu-24.04-arm`
    - `darwin/arm64` → `macos-latest`
  - `runs-on` 使用 matrix 变量引用 runner 标签
  - 通过 tag 提取 version（去掉 `v` 前缀），用于产物命名
  - 原生编译输出二进制名 `reddit-cli`（无需设置 `GOOS`/`GOARCH`）
  - 将二进制打包为 `reddit-cli-<version>-<os>-<arch>.tar.gz`
  - 使用 `actions/upload-artifact` 上传产物供后续 job 使用
- 风险/注意点: `ubuntu-24.04-arm` 是 GitHub 提供的 ARM64 Linux runner，公开仓库可用；私有仓库需确认是否可用

### ✅ 3. 添加 release job，创建 Release 并上传产物和校验和

- 优先级: P0
- 依赖项: 2
- 涉及文件: `.github/workflows/release.yml`
- 验收标准:
  - `release` job 依赖 `build` job，运行在 `ubuntu-latest`
  - 使用 `actions/download-artifact` 下载所有构建产物
  - 生成 `checksums.txt` 文件（SHA256）
  - 使用 `softprops/action-gh-release` 或 `gh release create` 创建 GitHub Release
  - 所有 `.tar.gz` 文件和 `checksums.txt` 均上传到 Release 页面
  - 使用 `GITHUB_TOKEN` 进行鉴权，无需额外 secrets
- 风险/注意点: `permissions` 需配置 `contents: write` 以允许创建 Release

## 实现建议

- 整个流水线在单个文件 `.github/workflows/release.yml` 中实现，分为 `test` → `build` → `release` 三个 job，通过 `needs` 串联
- version 提取推荐方式：`echo "${GITHUB_REF_NAME#v}"` 去掉 `v` 前缀
- matrix 中为每个目标定义 `goos`、`goarch`、`runner` 三个变量，`runs-on` 引用 `${{ matrix.runner }}`
- 原生编译直接 `go build -o reddit-cli .`，无需设置 GOOS/GOARCH 环境变量
- 使用 `actions/upload-artifact` + `actions/download-artifact` 在 job 间传递构建产物
- 校验和生成在 release job（Linux runner）中执行：`sha256sum *.tar.gz > checksums.txt`
