# 需求

## 目标与背景

为 `reddit-cli` 项目配置 GitHub Actions CI/CD 流水线，实现打 tag 后自动构建多平台二进制并发布到 GitHub Releases 页面，简化发版流程。

## 功能需求列表

### 核心功能

- 监听 tag 推送事件（格式 `v*`，如 `v1.0.0`）触发流水线
- 构建前运行 `go vet ./...` 和 `go test ./...`，任一失败则阻止发布
- 在对应平台的原生 runner 上编译二进制文件（不使用交叉编译）：
  - `linux/amd64` → `ubuntu-latest`
  - `linux/arm64` → `ubuntu-24.04-arm`
  - `darwin/arm64` → `macos-latest`（Apple Silicon）
- 自动创建与 tag 对应的 GitHub Release
- 将编译产物打包为 `.tar.gz` 格式并上传到 Release 页面
- 产物命名规则：`reddit-cli-<version>-<os>-<arch>.tar.gz`，其中 `<version>` 为 tag 去掉 `v` 前缀后的值（如 tag `v1.0.0` 对应产物 `reddit-cli-1.0.0-linux-amd64.tar.gz`）

### 扩展功能

- 生成 SHA256 校验和文件（`checksums.txt`），一并上传到 Release

## 非功能需求

- **构建环境**：使用 GitHub-hosted runner，各平台使用对应架构的 runner 原生编译；release job 使用 `ubuntu-latest`
- **Go 版本**：与项目 `go.mod` 保持一致（当前 go 1.26.3）
- **可维护性**：workflow 文件结构清晰，使用 matrix strategy 管理多平台构建，便于后续新增平台
- **安全**：使用 GitHub 自动提供的 `GITHUB_TOKEN`，无需额外配置 secrets

## 边界与不做事项

- 不支持 Windows 平台
- 不支持 macOS Intel（x86_64）平台
- 不做 Docker 镜像构建与发布
- 不做代码签名（code signing）

## 假设与约束

- **技术假设**：项目根目录 `main.go` 为编译入口（`go build -o reddit-cli .`）
- **Tag 格式**：使用语义化版本 `v*` 格式（如 `v1.0.0`、`v0.1.0-beta`）
- **环境约束**：仓库为 GitHub 托管，Actions 功能已启用

## 待确认事项

无
