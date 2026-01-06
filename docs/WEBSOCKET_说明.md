# WebSocket 支持说明

## 功能概述

simple-proxy 现在已经支持 WebSocket 连接的代理转发功能。

## 工作原理

### 自动检测
代理服务器会自动检测 WebSocket 升级请求（通过检查 `Upgrade: websocket` 头），并使用专门的处理逻辑。

### 连接流程

1. **检测 WebSocket 请求**
   - 检查 `Upgrade: websocket` 和 `Connection: upgrade` 头
   - 如果是 WebSocket 请求，使用 `handleWebSocket` 处理

2. **建立上游连接**
   - 支持通过 SOCKS5 代理连接上游 WebSocket 服务器
   - 支持通过 HTTP/HTTPS 代理（使用 CONNECT 方法）
   - 支持直连
   - 自动判断 ws 或 wss（基于上游端口 443）
   - 对于 wss，自动进行 TLS 包装和握手

3. **双向数据转发**
   - 使用 HTTP Hijacker 接管底层 TCP 连接
   - 同时进行客户端 ↔ 上游服务器的双向数据转发
   - 任一方向连接关闭时，自动结束转发

## 支持的代理类型

### SOCKS5 代理（优先级最高）
```bash
export SOCKS_PROXY="socks5://127.0.0.1:1080"
# 或带认证
export SOCKS_PROXY="socks5://username:password@127.0.0.1:1080"
```

### HTTP/HTTPS 代理
```bash
export HTTP_PROXY="http://127.0.0.1:8118"
export HTTPS_PROXY="http://127.0.0.1:8118"
```

### 直连
如果未设置任何代理环境变量，将直接连接到上游服务器。

## 协议支持

- **ws://** - 普通 WebSocket（上游端口非 443）
- **wss://** - 安全 WebSocket（上游端口为 443）

代理会根据 `UPSTREAM_HOST` 的端口自动判断使用 ws 还是 wss 协议。

## 使用示例

### 示例 1: 通过 SOCKS5 代理转发 WebSocket

```bash
export UPSTREAM_HOST="example.com:443"  # wss
export SOCKS_PROXY="socks5://127.0.0.1:1080"
export LISTEN_PORT="8080"
./simple-proxy
```

客户端连接：
```javascript
// 客户端连接到代理
const ws = new WebSocket('ws://localhost:8080/ws/path');
// 实际会被代理到 wss://example.com:443/ws/path
```

### 示例 2: 通过 HTTP 代理转发 WebSocket

```bash
export UPSTREAM_HOST="example.com:80"  # ws
export HTTP_PROXY="http://127.0.0.1:8118"
export LISTEN_PORT="8080"
./simple-proxy
```

### 示例 3: 直连转发 WebSocket

```bash
export UPSTREAM_HOST="example.com:443"
export LISTEN_PORT="8080"
./simple-proxy
```

## 日志输出

WebSocket 连接会产生以下日志：

```
===收到请求开始=== GET /ws/chat (Host: localhost:8080)
WebSocket 请求: /ws/chat (Host: localhost:8080)
WebSocket 连接建立，开始双向转发
WebSocket 连接关闭
```

## 技术细节

### TLS 配置
- 对于 wss 连接，自动使用 TLS 包装
- 设置 `InsecureSkipVerify: true`（适用于 IP 地址访问）
- 正确设置 `ServerName` 用于 TLS SNI

### 连接管理
- 使用 `net.Hijacker` 接口接管底层 TCP 连接
- 使用 `io.Copy` 进行高效的双向数据转发
- 使用 goroutine 并发处理双向数据流
- 通过 channel 监控连接状态

### 错误处理
- 自动识别常见的连接关闭错误（EOF, broken pipe, connection reset）
- 优雅地处理连接中断
- 详细的错误日志记录

## 注意事项

1. **Host 头保留**：代理会保留原始的 Host 头，确保上游服务器能正确识别虚拟主机

2. **端口判断**：自动根据上游端口判断使用 ws 还是 wss
   - 端口 443 → wss
   - 其他端口 → ws

3. **代理优先级**：SOCKS5 > HTTP/HTTPS > 直连

4. **超时设置**：连接建立超时为 10 秒

5. **并发安全**：每个 WebSocket 连接独立处理，互不干扰

## 测试建议

### 简单测试
可以使用以下工具测试 WebSocket 功能：

1. **wscat** (Node.js 工具)
```bash
npm install -g wscat
wscat -c ws://localhost:8080/ws/path
```

2. **浏览器控制台**
```javascript
const ws = new WebSocket('ws://localhost:8080/ws/test');
ws.onopen = () => console.log('连接已建立');
ws.onmessage = (e) => console.log('收到消息:', e.data);
ws.send('Hello Server');
```

3. **Python 测试**
```python
import websocket

ws = websocket.WebSocket()
ws.connect("ws://localhost:8080/ws/test")
ws.send("Hello")
result = ws.recv()
print(result)
ws.close()
```

## 性能特点

- **零拷贝**：使用 `io.Copy` 在内核层面进行数据传输
- **全双工**：真正的双向并发传输
- **低延迟**：直接的 TCP 层转发，无额外解析开销
- **连接复用**：支持多个并发 WebSocket 连接

## 与 HTTP 请求的区别

| 特性 | HTTP/HTTPS | WebSocket |
|------|-----------|-----------|
| 处理方式 | HTTP 客户端转发 | TCP 层直接转发 |
| 连接类型 | 短连接（或 Keep-Alive） | 长连接 |
| 数据流向 | 请求-响应 | 全双工 |
| 协议升级 | 不需要 | 需要 Upgrade 握手 |
| Hijacking | 不需要 | 需要接管连接 |



