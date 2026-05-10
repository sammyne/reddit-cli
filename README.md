# reddit-cli

Reddit in your terminal — Go 实现，对标 [public-clis/rdt-cli](https://github.com/public-clis/rdt-cli) v0.4.1。

## 快速开始

### 安装

```bash
# 从源码构建（需要 Go 1.21+）
git clone https://github.com/sammyne/reddit-cli.git
cd reddit-cli
go build -o reddit .

# 可选：移到 PATH 中
mv reddit /usr/local/bin/
```

### 1. 登录

在浏览器中登录 [reddit.com](https://www.reddit.com)，然后运行：

```bash
reddit login
# ✅ Login successful! (12 cookies extracted)
```

CLI 会自动从 Chrome/Firefox/Edge/Brave 中提取 session cookies，无需手动输入密码。

### 2. 检查状态

```bash
reddit status
# ✅ Authenticated (12 cookies)
#   user: your_username
#   capabilities: read, write
#   source: browser
```

支持结构化输出：

```bash
reddit status --json
reddit status --yaml
```

### 3. 查看个人资料

```bash
reddit whoami
# ╭─── 👤 Me ───╮
# │ u/your_username
# │ 📊 Total karma: 1.2K
# │    Post: 800 · Comment: 400
# │ 📅 Joined: 2020-03-15
# ╰──────────────╯
```

### 4. 搜索

```bash
# 基本搜索
reddit search "golang concurrency"

# 限定 subreddit + 排序 + 时间范围
reddit search "async await" -r programming --sort top --time year

# 限制结果数量
reddit search "rust vs go" -n 10

# JSON 输出（适合管道处理）
reddit search "python tips" --json

# 精简字段输出（适合 AI Agent）
reddit search "machine learning" --compact

# 导出到文件
reddit search "kubernetes" -o results.json
```

### 5. 登出

```bash
reddit logout
# ✅ Credentials cleared
```

## 命令一览

| 命令 | 说明 |
|------|------|
| `reddit login` | 从浏览器提取 Reddit cookies |
| `reddit logout` | 清除已保存的凭据 |
| `reddit status` | 检查认证状态 |
| `reddit whoami` | 显示当前用户资料 |
| `reddit search <query>` | 搜索 Reddit 帖子 |

### search 选项

| 选项 | 说明 |
|------|------|
| `-r, --subreddit` | 限定搜索范围到指定 subreddit |
| `-s, --sort` | 排序：`relevance` / `hot` / `top` / `new` / `comments` |
| `-t, --time` | 时间过滤：`hour` / `day` / `week` / `month` / `year` / `all` |
| `-n, --limit` | 结果数量（默认 25） |
| `--after` | 分页游标 |
| `--json` / `--yaml` | 结构化输出（带信封格式） |
| `-c, --compact` | 精简字段输出 |
| `--full-text` | 不截断标题 |
| `-o, --output` | 输出到文件 |

## 输出格式

- **终端（TTY）**：彩色表格渲染
- **管道/非 TTY**：自动切换为 YAML
- `--json` / `--yaml`：强制结构化输出，使用标准信封格式：

```json
{
  "ok": true,
  "schema_version": "1",
  "data": { ... }
}
```

- 环境变量 `OUTPUT=json|yaml|rich` 可覆盖默认行为

## 配置

凭据保存在 `$HOME/.config/reddit/credential.json`（回退到 `$HOME/.config/rdt-cli/`），文件权限 `0600`。

Cookies 有效期 7 天，过期后会自动尝试从浏览器重新提取。

## License

Apache-2.0
