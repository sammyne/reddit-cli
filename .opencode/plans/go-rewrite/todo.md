# TODO

## 任务列表

### ✅ 1. 项目脚手架：go mod init + 目录结构 + main.go + root 命令

- 优先级: P0
- 依赖项: 无
- 涉及文件:
  - `go.mod`（新建）
  - `cmd/reddit/main.go`（新建）
  - `internal/cmd/root.go`（新建）
- 验收标准: `go build ./cmd/reddit` 编译通过，`./reddit --version` 输出版本号 `0.4.1`，`./reddit --help` 显示帮助
- 风险/注意点: module 路径为 `github.com/sammyne/rdt-cli`；cobra 依赖需要 `go get`

### ✅ 2. 常量与配置：URL、请求头、排序选项、配置目录路径、RuntimeConfig

- 优先级: P0
- 依赖项: 1
- 涉及文件:
  - `internal/reddit/constants.go`（新建）
  - `internal/reddit/config.go`（新建）
- 验收标准: 包含 BASE_URL、SEARCH_URL、SUBREDDIT_SEARCH_URL、ME_URL 等端点常量；SEARCH_SORT_OPTIONS、TIME_FILTERS 等选项；ConfigDir() 函数实现优先 `$HOME/.config/reddit/` 回退 `$HOME/.config/rdt-cli/` 逻辑；RuntimeConfig 结构体定义超时和延迟参数
- 风险/注意点: ConfigDir 回退逻辑需判断目录是否存在

### ✅ 3. 自定义错误类型与错误码映射

- 优先级: P0
- 依赖项: 1
- 涉及文件:
  - `internal/reddit/errors.go`（新建）
- 验收标准: 定义 RedditApiError、SessionExpiredError、RateLimitError、NotFoundError、ForbiddenError；实现 ErrorCodeFor(err) 返回稳定错误码字符串（not_authenticated、rate_limited、not_found、forbidden、api_error）
- 风险/注意点: 用 Go 标准 error 接口和 errors.Is/As 模式

### ✅ 4. 数据模型：Post、ListingPage、UserProfile

- 优先级: P0
- 依赖项: 1
- 涉及文件:
  - `internal/reddit/models.go`（新建）
- 验收标准: 三个结构体定义，带 `json` tag；Post 包含 id/name/title/subreddit/author/score/num_comments/created_utc/permalink/url/selftext/is_self/over_18/is_video/stickied；ListingPage 包含 Items/After/Before；UserProfile 包含 name/link_karma/comment_karma/created_utc/is_gold/is_mod
- 风险/注意点: JSON tag 需与 Reddit API 字段名对齐

### ✅ 5. JSON 解析器 + 单元测试

- 优先级: P0
- 依赖项: 4
- 涉及文件:
  - `internal/reddit/parser.go`（新建）
  - `internal/reddit/parser_test.go`（新建）
- 验收标准: ParsePost 从 map 解析 Post；ParseListing 从 Reddit listing JSON 解析 ListingPage；ParseUserProfile 解析用户资料；表驱动测试覆盖正常和边界情况（空字段、缺失字段、类型错误）；`go test ./internal/reddit/` 通过
- 风险/注意点: Reddit API 返回的字段类型不稳定（score 可能是 int 或 float），需要做安全类型转换

### ✅ 6. 浏览器指纹：平台感知的请求头生成

- 优先级: P0
- 依赖项: 2
- 涉及文件:
  - `internal/reddit/fingerprint.go`（新建）
- 验收标准: BrowserFingerprint 结构体；NewBrowserFingerprint() 根据 runtime.GOOS 生成对应的 User-Agent 和 sec-ch-ua-platform（macOS/Linux/Windows）；ReadHeaders() 返回完整读请求头 map
- 风险/注意点: 无

### ✅ 7. Session 管理：SessionState 与能力检测

- 优先级: P0
- 依赖项: 3
- 涉及文件:
  - `internal/reddit/session.go`（新建）
- 验收标准: SessionState 结构体（cookies/source/username/modhash/capabilities 等）；FromCredential() 构造函数；RefreshCapabilities() 检测 read/write 能力；ApplyIdentity() 从 /api/me.json 响应更新 username 和 modhash；IsAuthenticated()/CanWrite() 便捷方法
- 风险/注意点: 无

### ✅ 8. 凭据管理：Credential + 持久化 + 浏览器 cookie 提取

- 优先级: P0
- 依赖项: 2, 7
- 涉及文件:
  - `internal/reddit/auth.go`（新建）
- 验收标准: Credential 结构体（cookies/source/username/modhash/saved_at/last_verified_at）；SaveCredential() 写入 JSON 文件（0600 权限）；LoadCredential() 带 7 天 TTL 检查，过期自动尝试浏览器刷新；ClearCredential() 删除文件；ExtractBrowserCredential() 用 kooky 库提取 .reddit.com cookies；GetCredential() 链式尝试：saved → browser → nil
- 风险/注意点: kooky 库的 API 需要验证；浏览器正在运行时 SQLite 可能被锁

### ✅ 9. HTTP 传输层：ReadTransport（jitter + 重试 + cookie 回写）

- 优先级: P0
- 依赖项: 2, 3, 6, 7
- 涉及文件:
  - `internal/reddit/transport.go`（新建）
- 验收标准: ReadTransport 结构体包裹 http.Client；Request() 方法实现 Gaussian jitter 延迟（均值 0.3s、5% 概率 2-5s 额外延迟）；429 返回时按 Retry-After 等待；5xx 返回时指数退避（最多 3 次）；401→SessionExpiredError、403→ForbiddenError、404→NotFoundError；响应 cookie 回写 session；请求计数 + log 日志；HTML 响应检测
- 风险/注意点: net/http 的 cookie jar 管理方式与 Python httpx 不同，需手动管理 cookie

### ✅ 10. Reddit API 客户端：RedditClient

- 优先级: P0
- 依赖项: 9
- 涉及文件:
  - `internal/reddit/client.go`（新建）
- 验收标准: RedditClient 结构体组合 ReadTransport + SessionState；实现 Open()/Close() 生命周期；GetMe() 请求 /api/me.json 并调用 session.ApplyIdentity()；ValidateSession() 探测认证状态返回结构化 map；Search() 支持 query/subreddit/sort/time_filter/limit/after 参数，请求 /search.json 或 /r/{sub}/search.json
- 风险/注意点: 注意 Search URL 中 restrict_sr 参数在有 subreddit 时为 "on"

### ✅ 11. 索引缓存：SaveIndex + GetItemByIndex

- 优先级: P1
- 依赖项: 2, 4
- 涉及文件:
  - `internal/reddit/indexcache.go`（新建）
- 验收标准: SaveIndex() 将 []Post 保存到 index_cache.json（含 source/saved_at/count/items）；GetItemByIndex(idx) 按 1-based 索引返回缓存条目；GetIndexInfo() 返回缓存元数据；文件权限 0600
- 风险/注意点: 为后续 show 命令预留，本阶段 search 命令会调用 SaveIndex

### ✅ 12. 输出辅助：信封格式、JSON/YAML 打印、格式化工具、TTY 检测

- 优先级: P0
- 依赖项: 3
- 涉及文件:
  - `internal/reddit/output.go`（新建）
- 验收标准: SuccessPayload(data) 包装 `{ok:true, schema_version:"1", data:...}`；ErrorPayload(code, message) 包装错误信封；PrintJSON()/PrintYAML() 输出到 stdout；MaybePrintStructured() 根据 flags/env/TTY 决定是否输出结构化格式并返回 bool；ResolveOutputFormat() 实现 flag→env→TTY 优先级；FormatScore()（1.2k 格式）；FormatTime()（相对时间）；EmitError() 结构化错误输出；SaveOutputToFile() 按扩展名自动选 JSON/YAML
- 风险/注意点: 非 TTY 默认 YAML；环境变量 OUTPUT 可覆盖

### ✅ 13. Auth 命令：login / logout / status / whoami

- 优先级: P0
- 依赖项: 8, 10, 12
- 涉及文件:
  - `internal/cmd/auth.go`（新建）
  - `internal/cmd/root.go`（修改，注册子命令）
- 验收标准:
  - `reddit login`：调用 ExtractBrowserCredential，成功显示 cookie 数量，失败提示去浏览器登录
  - `reddit logout`：调用 ClearCredential，显示确认
  - `reddit status [--json|--yaml]`：显示 authenticated/cookie_count/username/capabilities/modhash_present/source/error；有凭据时调用 ValidateSession 做活跃检查
  - `reddit whoami [--json|--yaml]`：调用 GetMe + GetUserAbout（本阶段用 GetMe 数据），显示用户面板（用户名/karma/注册时间/Gold/Mod）
  - 所有命令的终端输出带颜色
- 风险/注意点: status 需先尝试加载凭据，再做在线验证；whoami 需要 require_auth 守卫

### ✅ 14. Search 命令：reddit search

- 优先级: P0
- 依赖项: 5, 10, 11, 12
- 涉及文件:
  - `internal/cmd/search.go`（新建）
  - `internal/cmd/root.go`（修改，注册子命令）
- 验收标准:
  - `reddit search "python async"` 显示彩色表格（#/Score/Subreddit/Title/Author/Comments）
  - `-r programming --sort top --time year` 过滤正常工作
  - `-n 10` 限制结果数
  - `--after <cursor>` 分页正常，有下一页时显示提示命令
  - `--json`/`--yaml` 输出带信封格式
  - `--compact` 精简字段后默认 YAML 输出
  - `--full-text` 不截断标题
  - `-o results.json` 输出到文件
  - 搜索结果调用 SaveIndex 保存到索引缓存
  - 错误时输出结构化错误（JSON/YAML 模式）或彩色错误（终端模式）
- 风险/注意点: 分页游标 after 从 ListingPage 提取

### ✅ 15. 集成验证：go build + go vet + go test

- 优先级: P0
- 依赖项: 13, 14
- 涉及文件: 无新文件
- 验收标准: `go build ./cmd/reddit` 编译通过；`go vet ./...` 无警告；`go test ./...` 全部通过；`go mod tidy` 后 go.mod/go.sum 干净
- 风险/注意点: 确保所有 import 路径正确

## 实现建议

- 项目目录结构：
  ```
  cmd/reddit/main.go           # main 入口
  internal/
    reddit/                    # 核心库（单包，扁平结构）
      constants.go
      config.go
      errors.go
      models.go
      parser.go / parser_test.go
      fingerprint.go
      session.go
      auth.go
      transport.go
      client.go
      indexcache.go
      output.go
    cmd/                       # cobra 命令定义
      root.go
      auth.go
      search.go
  ```
- 参考源码在 `_reference/rdt_cli/` 目录，逐文件对照实现
- 优先使用标准库，仅引入 cobra、kooky、go-pretty、yaml.v3 四个第三方依赖
- parser_test.go 使用真实 Reddit API 响应片段作为测试数据（可从参考项目的 tests/ 目录获取灵感）
