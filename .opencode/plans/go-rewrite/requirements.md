# 需求

## 目标与背景

用 Go 1.24 重写 Python 项目 [public-clis/rdt-cli](https://github.com/public-clis/rdt-cli) v0.4.1 的**部分功能**（第一阶段）。原项目是一个终端 Reddit 客户端。本阶段只实现认证全部命令 + search 命令，后续阶段再补齐其余命令。

## 功能需求列表

### 本阶段实现的命令（共 5 个）

#### 1. 认证命令 (Auth) — 4 个
- `reddit login`：从本地浏览器（Chrome/Firefox/Edge/Brave）提取 Reddit session cookies，保存到本地
- `reddit logout`：清除已保存的 credential 文件
- `reddit status`：检查认证状态（cookie 数量、用户名、能力、modhash），支持 `--json`/`--yaml` 输出
- `reddit whoami`：显示当前用户详细资料（用户名、总 karma、帖子/评论 karma、注册时间、Gold/Mod 状态），支持 `--json`/`--yaml`

#### 2. 搜索命令 (Search) — 1 个
- `reddit search <query>`：搜索 Reddit 帖子
  - `-r/--subreddit`：限定 subreddit 范围
  - `-s/--sort`：排序（relevance/hot/top/new/comments）
  - `-t/--time`：时间过滤（hour/day/week/month/year/all）
  - `-n/--limit`：结果数量（默认 25）
  - `--after`：分页游标
  - `--json`/`--yaml`：结构化输出
  - `--compact`/`--full-text`/`-o`：列表输出选项

### 本阶段需要的基础设施

#### 3. Reddit API 客户端
- 基于 `https://www.reddit.com` 的 JSON API（非 OAuth）
- 本阶段只需 ReadTransport（GET 请求），不需要 WriteTransport
- Gaussian jitter 延迟（0.3s 均值 + 5% 概率 2-5s 长停顿）
- 指数退避重试（429/5xx，最多 3 次）
- 响应 cookie 回写到 session
- 请求计数与日志
- 需要实现的客户端方法：`GetMe()`、`ValidateSession()`、`Search()`

#### 4. 浏览器指纹
- Chrome 133 User-Agent，根据运行时 `runtime.GOOS` 动态生成平台部分：
  - macOS → `Macintosh; Intel Mac OS X 10_15_7`，`sec-ch-ua-platform: "macOS"`
  - Linux → `X11; Linux x86_64`，`sec-ch-ua-platform: "Linux"`
  - Windows → `Windows NT 10.0; Win64; x64`，`sec-ch-ua-platform: "Windows"`
- `sec-ch-ua`、`sec-ch-ua-mobile`、Sec-Fetch 系列头保持固定
- 读请求头（base_headers）

#### 5. Session 管理
- 从 Credential 构建 SessionState
- 能力检测：`read`（有 reddit_session cookie）、`write`（有 modhash）
- `ApplyIdentity()` 从 /api/me.json 更新用户名和 modhash
- `ValidateSession()` 探测认证状态

#### 6. 凭据管理 (Auth)
- 浏览器 cookie 提取（Chrome/Firefox/Edge/Brave，读取 SQLite cookie 数据库）
- 保存到 `$HOME/.config/reddit/credential.json`，权限 0600；若目录不存在则回退到 `$HOME/.config/rdt-cli/credential.json`
- 7 天 TTL，过期自动尝试浏览器刷新
- Credential 结构：cookies、source、username、modhash、saved_at、last_verified_at
- 加载/保存/清除操作

#### 7. 数据模型（本阶段子集）
- `Post`：id, name, title, subreddit, author, score, num_comments, created_utc, permalink, url, selftext, is_self, over_18, is_video, stickied
- `ListingPage`：items []Post, after, before
- `UserProfile`：name, link_karma, comment_karma, created_utc, is_gold, is_mod
- 均支持 JSON 序列化

#### 8. JSON 解析器（本阶段子集）
- `ParsePost`：从 Reddit JSON payload 解析单个 Post
- `ParseListing`：解析 listing 响应为 ListingPage
- `ParseUserProfile`：解析用户资料

#### 9. 索引缓存
- search 结果保存到 `$HOME/.config/reddit/index_cache.json`；若目录不存在则回退到 `$HOME/.config/rdt-cli/index_cache.json`
- 支持通过 1-based 短索引检索帖子（为后续 show 命令预留）
- 文件权限 0600

#### 10. 输出格式
- 终端彩色表格渲染（search 结果列表）
- 终端彩色面板渲染（whoami 用户信息、status 状态）
- `--json`：JSON + 信封格式 `{"ok": true, "schema_version": "1", "data": ...}`
- `--yaml`：YAML + 信封格式
- `--compact`：精简字段（search 命令）
- `--full-text`：不截断标题（search 命令）
- `-o/--output`：输出到文件
- 非 TTY 默认 YAML 输出
- 环境变量 `OUTPUT` 覆盖输出模式

#### 11. 错误处理
- 自定义错误类型：RedditApiError、SessionExpiredError、RateLimitError、NotFoundError、ForbiddenError
- 错误码映射（error_code_for_exception 等价物）
- 结构化错误输出（JSON/YAML 信封）
- SessionExpired 时自动尝试浏览器 cookie 刷新重试

#### 12. 常量
- API 端点 URL（HOME_URL、SEARCH_URL、SUBREDDIT_SEARCH_URL、ME_URL 等本阶段用到的）
- 请求头（Chrome 133 macOS）
- 排序/时间过滤选项
- 配置目录路径：优先 `$HOME/.config/reddit/`，回退 `$HOME/.config/rdt-cli/`

### 本阶段不做的事项
- Browse 命令组（feed/popular/all/sub/sub-info/user/user-posts/user-comments/saved/upvoted/open）
- Post 命令组（read/show）
- Social 命令组（upvote/save/subscribe/comment）
- Export 命令
- WriteTransport（写请求传输层）
- Comment 模型和评论树解析
- SubredditInfo 模型

## 非功能需求
- **安全**：credential 文件权限 0600；不在日志中输出 cookie 值
- **兼容性**：macOS 优先，Linux/Windows 兼顾
- **可维护性**：遵循 AGENTS.md 中的 Go 编码规范；函数不超过约 50 行
- **测试要求**：parser 解析逻辑编写单元测试（表驱动）

## 假设与约束
- **CLI 框架**：`cobra`（Go 生态标准）
- **HTTP 客户端**：标准库 `net/http`
- **YAML 输出**：`gopkg.in/yaml.v3`
- **浏览器 Cookie 提取**：使用 `github.com/browserutils/kooky` 或类似库读取浏览器 cookie 数据库
- **终端渲染**：使用 `github.com/jedib0t/go-pretty` 做表格和彩色输出
- **二进制名称**：`reddit`（已确认）
- **Go module 路径**：`github.com/sammyne/rdt-cli`

## 待确认事项
- 无
