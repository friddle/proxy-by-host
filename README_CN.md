# Reserver HTTPS Proxy (中文文档)

这是一个轻量级、安全的正向代理工具，旨在帮助团队通过简单修改本地 DNS/Hosts 即可访问受限的上游资源（例如通过 VPN 或上游代理访问）。

## 核心应用场景

将此工具部署在已连接 VPN 或拥有特殊网络访问权限的服务器上。普通用户无需在自己的电脑上安装 VPN 客户端或配置复杂的代理设置，只需将特定域名（如 `google.com`）的 DNS 解析指向该服务器 IP。

当用户访问这些域名时，请求会发送到该服务器，由 `reserver-proxy` 接收并通过配置的上游代理转发到真实目标地址。

**主要特点：**
*   **零配置客户端：** 用户仅需修改 Hosts 文件。
*   **支持上游代理：** 支持 HTTP/HTTPS/SOCKS5 上游代理链。
*   **自动 SSL：** 支持自动生成自签名证书（用于测试）。
*   **HTTPS 支持：** 监听 443 端口并处理 TLS（注意 SSL 证书信任问题）。

## 快速开始

### 1. 安装

从 [Releases](https://github.com/friddle/reserver-https-proxy/releases) 下载二进制文件，或自行编译：

```bash
go build -o reserver-proxy
```

### 2. 使用方法

**基础用法 (默认端口 80/443):**

```bash
sudo ./reserver-proxy --ssl=generate
```

**配合上游代理 (翻墙模式):**
假设服务器上运行着一个代理客户端（如 Clash/V2Ray），监听在 7897 端口。

```bash
sudo ./reserver-proxy --ssl=generate --proxy=http://127.0.0.1:7897
```

**自定义端口和证书:**

```bash
sudo ./reserver-proxy \
  --http-port=8080 \
  --https-port=8443 \
  --ssl=on \
  --ssl-crt=/path/to/cert.crt \
  --ssl-key=/path/to/key.key
```

### 3. 客户端配置 (DNS 欺骗)

在用户电脑上，编辑 Hosts 文件 (`/etc/hosts` 或 Windows 的 `System32\drivers\etc\hosts`)：

```text
# 将目标域名指向运行 reserver-proxy 的服务器 IP
192.168.1.100 www.google.com
192.168.1.100 google.com
```

此时，访问 `https://www.google.com` 的流量会直接发往 `192.168.1.100`，代理服务器会根据 Host 头将其转发出去。

## 命令行参数

| 参数 | 默认值 | 说明 |
|------|---------|-------------|
| `--ssl` | `none` | SSL 模式: `generate` (自动生成), `on` (使用指定证书), `none` (关闭). |
| `--ssl-crt` | - | SSL 证书路径 (当 `--ssl=on` 时必须). |
| `--ssl-key` | - | SSL 密钥路径 (当 `--ssl=on` 时必须). |
| `--proxy` | - | 上游代理地址 (例如 `http://127.0.0.1:7897`). |
| `--http-port`| `80` | HTTP 监听端口. |
| `--https-port`| `443` | HTTPS 监听端口. |

## 关于 SSL/HTTPS 证书的说明

当使用 `--ssl=generate` 模式时，服务器会自动生成一个针对 `*.reserver.proxy` 的自签名证书。

如果你通过 DNS 欺骗方式访问真实域名（如 `google.com`）：
1.  **浏览器警告：** 浏览器会提示“连接不安全”，因为代理提供的证书（`*.reserver.proxy`）与访问的域名（`google.com`）不匹配。
2.  **处理方式：** 你需要在浏览器中点击“高级” -> “继续访问”。
3.  **HSTS 限制：** 对于开启了严格 HSTS 的网站（如 Google），浏览器可能会完全阻止访问，无法跳过警告。

**生产环境建议：** 如果需要长期稳定使用，建议为目标域名申请合法的 SSL 证书，并使用 `--ssl=on` 加载该证书。

## 许可证

MIT

