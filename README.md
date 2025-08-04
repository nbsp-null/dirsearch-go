# dirsearch-go

一个用Go语言重写的dirsearch工具，用于Web路径暴力破解。

## 功能特性

- 🚀 高性能并发扫描
- 📝 多种报告格式支持 (JSON, CSV, HTML, Plain Text)
- 🔧 灵活的配置选项
- 🎯 精确的路径过滤
- 🔄 递归扫描支持
- 🌐 代理支持
- 📊 实时进度显示

## 安装

### 从源码编译

```bash
# 克隆仓库
git clone https://github.com/your-username/dirsearch-go.git
cd dirsearch-go

# 安装依赖
go mod tidy

# 编译
go build -o dirsearch-go main.go
```

### 使用预编译版本

从 [Releases](https://github.com/your-username/dirsearch-go/releases) 页面下载适合你系统的预编译版本。

## 使用方法

### 基本用法

```bash
# 扫描单个目标
./dirsearch-go -u https://example.com -w wordlist.txt

# 扫描多个目标
./dirsearch-go -u https://example.com -u https://test.com -w wordlist.txt

# 使用扩展名
./dirsearch-go -u https://example.com -w wordlist.txt -e php,html,js

# 设置线程数
./dirsearch-go -u https://example.com -w wordlist.txt -t 50
```

### 高级用法

```bash
# 递归扫描
./dirsearch-go -u https://example.com -w wordlist.txt -r

# 使用代理
./dirsearch-go -u https://example.com -w wordlist.txt -p http://proxy:8080

# 自定义请求头
./dirsearch-go -u https://example.com -w wordlist.txt -H "Authorization: Bearer token"

# 保存报告
./dirsearch-go -u https://example.com -w wordlist.txt -o report.json --format json
```

## 命令行选项

### 必需参数

- `-u, --url`: 目标URL (可多次使用)
- `-w, --wordlists`: 字典文件路径

### 字典设置

- `-e, --extensions`: 扩展名列表 (如: php,asp)
- `-f, --force-extensions`: 强制添加扩展名到所有字典条目
- `-O, --overwrite-extensions`: 覆盖字典中的其他扩展名
- `--exclude-extensions`: 排除的扩展名列表
- `--remove-extensions`: 移除所有路径中的扩展名
- `--prefixes`: 添加到所有字典条目的前缀
- `--suffixes`: 添加到所有字典条目的后缀
- `-U, --uppercase`: 大写字典
- `-L, --lowercase`: 小写字典
- `-C, --capital`: 首字母大写字典

### 通用设置

- `-t, --threads`: 线程数 (默认: 25)
- `--async`: 启用异步模式
- `-r, --recursive`: 递归暴力破解
- `--deep-recursive`: 在每个目录深度执行递归扫描
- `--force-recursive`: 对所有找到的路径进行递归暴力破解
- `-R, --max-recursion-depth`: 最大递归深度
- `--recursion-status`: 执行递归扫描的有效状态码
- `--subdirs`: 扫描给定URL的子目录
- `--exclude-subdirs`: 递归扫描期间排除的子目录
- `-i, --include-status`: 包含的状态码
- `-x, --exclude-status`: 排除的状态码
- `--exclude-sizes`: 按大小排除响应
- `--exclude-text`: 按文本排除响应
- `--exclude-regex`: 按正则表达式排除响应
- `--exclude-redirect`: 按重定向URL排除响应
- `--exclude-response`: 按响应页面排除响应
- `--skip-on-status`: 遇到这些状态码时跳过目标
- `--min-response-size`: 最小响应长度
- `--max-response-size`: 最大响应长度
- `--max-time`: 扫描的最大运行时间
- `--exit-on-error`: 发生错误时退出

### 请求设置

- `-m, --http-method`: HTTP方法 (默认: GET)
- `-d, --data`: HTTP请求数据
- `--data-file`: 包含HTTP请求数据的文件
- `-H, --header`: HTTP请求头 (可多次使用)
- `--headers-file`: 包含HTTP请求头的文件
- `-F, --follow-redirects`: 跟随HTTP重定向
- `--random-agent`: 为每个请求选择随机User-Agent
- `--auth`: 认证凭据
- `--auth-type`: 认证类型
- `--cert-file`: 包含客户端证书的文件
- `--key-file`: 包含客户端证书私钥的文件
- `--user-agent`: User-Agent
- `--cookie`: Cookie

### 连接设置

- `--timeout`: 连接超时
- `--delay`: 请求之间的延迟
- `-p, --proxy`: 代理URL (可多次使用)
- `--proxies-file`: 包含代理服务器的文件
- `--proxy-auth`: 代理认证凭据
- `--replay-proxy`: 用于重放找到路径的代理
- `--tor`: 使用Tor网络作为代理
- `--scheme`: 原始请求或URL中没有方案时的方案
- `--max-rate`: 每秒最大请求数
- `--retries`: 失败请求的重试次数
- `--ip`: 服务器IP地址
- `--interface`: 要使用的网络接口

### 高级设置

- `--crawl`: 在响应中爬取新路径

### 视图设置

- `--full-url`: 输出中的完整URL
- `--redirects-history`: 显示重定向历史
- `--no-color`: 无彩色输出
- `-q, --quiet-mode`: 静默模式

### 输出设置

- `-o, --output`: 输出文件或MySQL/PostgreSQL URL
- `--format`: 报告格式 (可用: simple, plain, json, xml, md, csv, html, sqlite, mysql, postgresql)
- `--log`: 日志文件

## 配置文件

dirsearch-go 支持配置文件。默认配置文件为 `config.ini`，也可以通过 `--config` 参数指定。

配置文件示例:

```ini
[general]
threads = 25
async = False
recursive = False
deep-recursive = False
force-recursive = False
recursion-status = 200-399,401,403
max-recursion-depth = 0
exclude-subdirs = %ff/,.,;/,..;/,;/,./,../,%2e/,%2e%2e/
random-user-agents = False
max-time = 0
exit-on-error = False

[dictionary]
default-extensions = php,aspx,jsp,html,js
force-extensions = False
overwrite-extensions = False
lowercase = False
uppercase = False
capitalization = False

[request]
http-method = get
follow-redirects = False

[connection]
timeout = 7.5
delay = 0
max-rate = 0
max-retries = 1

[advanced]
crawl = False

[view]
full-url = False
quiet-mode = False
color = True
show-redirects-history = False

[output]
report-format = plain
autosave-report = True
autosave-report-folder = reports/
```

## 报告格式

dirsearch-go 支持多种报告格式:

- **plain**: 纯文本格式，包含详细信息
- **simple**: 简单格式，只显示状态码和路径
- **json**: JSON格式，便于程序处理
- **csv**: CSV格式，便于在电子表格中查看
- **html**: HTML格式，包含样式和表格

## 示例

### 基本扫描

```bash
./dirsearch-go -u https://example.com -w /path/to/wordlist.txt
```

### 使用扩展名

```bash
./dirsearch-go -u https://example.com -w wordlist.txt -e php,html,js
```

### 递归扫描

```bash
./dirsearch-go -u https://example.com -w wordlist.txt -r --max-recursion-depth 3
```

### 使用代理

```bash
./dirsearch-go -u https://example.com -w wordlist.txt -p http://127.0.0.1:8080
```

### 保存报告

```bash
./dirsearch-go -u https://example.com -w wordlist.txt -o report.json --format json
```

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

本项目采用 GNU General Public License v2.0 许可证。

## 致谢

本项目基于 [dirsearch](https://github.com/maurosoria/dirsearch) 项目重写，感谢原作者的贡献。 