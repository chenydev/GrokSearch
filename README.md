# Grok Search CLI

Grok Search CLI 是一个轻量 Go 命令行工具，用于通过 Grok 兼容接口搜索网页、通过 Tavily 抓取网页/映射站点，并在需要时使用 Firecrawl 作为抓取兜底。

它面向终端、脚本和 AI 客户端 skill 集成，不需要运行 MCP 服务。

## 安装与构建

本地构建：

```bash
CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o dist/grok-search ./cmd/grok-search
```

运行测试：

```bash
go test ./...
go vet ./...
```

如果你使用 NixOS，可以在自己的环境中提供 Go；仓库不会提交 `.envrc`，避免强绑定个人 Nix 配置。

## 配置

配置优先级：

1. 环境变量
2. `--config` 指定的 JSON 配置文件，或默认用户配置文件
3. 内置默认值

默认配置路径使用 Go 的 `os.UserConfigDir()`：

- Linux: `$XDG_CONFIG_HOME/grok-search/config.json` 或 `~/.config/grok-search/config.json`
- macOS: `~/Library/Application Support/grok-search/config.json`
- Windows: `%AppData%\grok-search\config.json`

配置文件示例：

```json
{
  "guda_api_key": "",
  "guda_base_url": "https://code.guda.studio",
  "grok_api_url": "",
  "grok_api_key": "",
  "grok_model": "grok-4.20-fast",
  "tavily_enabled": true,
  "tavily_api_url": "https://api.tavily.com",
  "tavily_api_key": "",
  "firecrawl_api_url": "https://api.firecrawl.dev/v2",
  "firecrawl_api_key": "",
  "debug": false,
  "retry_max_attempts": 3,
  "retry_multiplier": 1,
  "retry_max_wait": 10
}
```

支持的环境变量：

```text
GUDA_API_KEY
GUDA_BASE_URL
GROK_API_URL
GROK_API_KEY
GROK_MODEL
TAVILY_ENABLED
TAVILY_API_URL
TAVILY_API_KEY
FIRECRAWL_API_URL
FIRECRAWL_API_KEY
GROK_DEBUG
GROK_RETRY_MAX_ATTEMPTS
GROK_RETRY_MULTIPLIER
GROK_RETRY_MAX_WAIT
```

如果设置 `GUDA_API_KEY`，且没有显式设置 `GROK_*`、`TAVILY_*`、`FIRECRAWL_*`，CLI 会从 `GUDA_BASE_URL` 自动派生服务地址和 key。

## 常用命令

检查配置：

```bash
grok-search config get
grok-search config check
```

列出模型：

```bash
grok-search models
```

搜索：

```bash
grok-search search "FastAPI latest dependency injection docs"
```

搜索并输出 JSON：

```bash
grok-search search "FastAPI latest dependency injection docs" --format json
```

搜索并单独保存信源：

```bash
grok-search search "FastAPI latest dependency injection docs" --sources sources.json
```

抓取网页：

```bash
grok-search fetch https://example.com/page
```

映射站点：

```bash
grok-search map https://docs.example.com --instructions "only docs"
```

查看最近一次搜索缓存：

```bash
grok-search last --format json
```

## GitHub Actions 产物

仓库提供 `.github/workflows/build.yml`：

- Pull Request 和 push 时会运行测试并构建多平台二进制
- tag `v*` 推送时会创建 GitHub Release 并上传压缩产物

支持目标：

- Linux amd64 / arm64
- macOS amd64 / arm64
- Windows amd64 / arm64
