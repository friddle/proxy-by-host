# Simple Proxy

一个简单的 Go 反向代理服务，用于将请求转发到固定 IP 地址，并通过代理访问。

## 功能特点

- ✅ 接收任意 Host 的 HTTP 请求
- ✅ 将请求转发到固定 IP 地址（保持原始 Host 头）
- ✅ 支持通过 HTTP/HTTPS 代理访问上游服务器
- ✅ 支持通过 SOCKS5 代理访问上游服务器
- ✅ 自动处理请求头和响应头
- ✅ 支持环境变量配置

## 使用场景

当你需要：
- 访问某个域名（如 webmail.prod.code27.cn），但 DNS 解析不可用或不正确
- 手动指定该域名对应的真实 IP 地址
- 并且需要通过代理（如科学上网工具）访问该 IP

这个代理服务可以帮你实现类似 `curl -H "Host: webmail.prod.code27.cn" http://47.252.16.154` 但通过代理访问的效果。

## 快速开始

### 方法一：直接运行

1. 安装依赖：
```bash
go mod download
```

2. 设置环境变量并运行：
```bash
# 使用 HTTP 代理
export UPSTREAM_HOST=47.252.16.154
export HTTP_PROXY=http://127.0.0.1:7890
export HTTPS_PROXY=http://127.0.0.1:7890
export LISTEN_PORT=8080

go run main.go
```

或者使用 SOCKS5 代理：
```bash
export UPSTREAM_HOST=47.252.16.154
export SOCKS_PROXY=socks5://127.0.0.1:1080
export LISTEN_PORT=8080

go run main.go
```

3. 测试：
```bash
# 访问代理服务
curl -H "Host: webmail.prod.code27.cn" http://localhost:8080/

# 或者直接用域名（需要配置 hosts 或 DNS 指向 localhost）
curl http://webmail.prod.code27.cn:8080/
```

### 方法二：使用 Docker

1. 构建镜像：
```bash
docker build -t simple-proxy .
```

2. 运行容器：
```bash
docker run -d \
  -p 8080:8080 \
  -e UPSTREAM_HOST=47.252.16.154 \
  -e HTTP_PROXY=http://host.docker.internal:7890 \
  --name simple-proxy \
  simple-proxy
```

### 方法三：使用 Docker Compose

1. 复制环境变量配置：
```bash
cp .env.example .env
```

2. 编辑 `.env` 文件，设置你的代理配置

3. 启动服务：
```bash
docker-compose up -d
```

4. 查看日志：
```bash
docker-compose logs -f
```

## 环境变量配置

| 变量名 | 说明 | 默认值 | 示例 |
|--------|------|--------|------|
| `UPSTREAM_HOST` | 上游服务器的 IP 地址 | `47.252.16.154` | `47.252.16.154` |
| `LISTEN_PORT` | 代理服务监听端口 | `8080` | `8080` |
| `HTTP_PROXY` | HTTP 代理地址 | - | `http://127.0.0.1:7890` |
| `HTTPS_PROXY` | HTTPS 代理地址 | - | `http://127.0.0.1:7890` |
| `SOCKS_PROXY` | SOCKS5 代理地址 | - | `socks5://127.0.0.1:1080` |

**注意**：
- SOCKS_PROXY 优先级高于 HTTP_PROXY
- 如果都不设置，将直接连接（不推荐）
- 代理地址支持认证：`socks5://user:pass@host:port` 或 `http://user:pass@host:port`

## 工作原理

```
客户端请求
   ↓
   → [本代理服务:8080]
        ↓
        → 保持原始 Host 头
        → 修改目标地址为固定 IP
        → 通过代理发送请求
           ↓
           → [HTTP/SOCKS代理]
                ↓
                → [上游服务器 47.252.16.154]
                     ↓
                     ← 返回响应
                ↓
           ← 通过代理返回
        ↓
   ← 返回给客户端
```

## 实际应用示例

假设你在中国，需要访问 `webmail.prod.code27.cn`，该域名解析到 `47.252.16.154`，但你需要通过代理访问：

1. 启动代理服务：
```bash
export UPSTREAM_HOST=47.252.16.154
export HTTP_PROXY=http://127.0.0.1:7890  # 你的代理地址
go run main.go
```

2. 配置你的应用或浏览器：
   - 方式一：修改 `/etc/hosts`，将域名指向 `127.0.0.1`
   - 方式二：直接使用代理服务的地址访问

3. 现在访问 `http://localhost:8080` 就会通过代理访问到真实的服务器

## 编译

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o simple-proxy-linux main.go

# macOS
GOOS=darwin GOARCH=amd64 go build -o simple-proxy-macos main.go

# Windows
GOOS=windows GOARCH=amd64 go build -o simple-proxy-windows.exe main.go
```

## 日志

程序会输出详细的请求日志：
```
2024/01/01 12:00:00 代理服务器启动在 :8080
2024/01/01 12:00:00 上游服务器: 47.252.16.154
2024/01/01 12:00:00 HTTP代理: http://127.0.0.1:7890
2024/01/01 12:00:01 收到请求: GET / (Host: webmail.prod.code27.cn)
2024/01/01 12:00:01 转发请求到: http://47.252.16.154/ (Host头: webmail.prod.code27.cn)
2024/01/01 12:00:02 请求完成: GET / -> 200 (1234 bytes)
```

## 故障排查

### 问题：连接被拒绝
- 检查 UPSTREAM_HOST 是否正确
- 检查代理设置是否正确
- 验证代理服务是否在运行

### 问题：SSL/TLS 错误
- 如果上游服务器使用 HTTPS，确保证书有效
- 可以在代码中临时设置 `InsecureSkipVerify: true`（仅用于测试）

### 问题：Host 头不匹配
- 程序会自动保持原始 Host 头
- 可以通过 `-H "Host: xxx"` 手动指定

## 许可证

MIT



