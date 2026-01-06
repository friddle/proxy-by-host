# Simple Proxy 快速开始

## 📦 镜像信息

```
镜像地址: registrylan.code27.cn/third/simple_proxy:latest
架构: linux/amd64
大小: 14.4MB
状态: ✅ 已推送成功
```

## 🚀 一键部署

### 1. 创建 docker-compose.yml

```bash
cat > docker-compose.yml << 'EOF'
version: '3.8'

services:
  simple-proxy:
    image: registrylan.code27.cn/third/simple_proxy:latest
    container_name: simple-proxy
    restart: unless-stopped
    ports:
      - "80:80"
    environment:
      UPSTREAM_HOST: "47.90.201.196:443"
      LISTEN_PORT: "80"
      HTTP_PROXY: "http://host.docker.internal:7897"
      HTTPS_PROXY: "http://host.docker.internal:7897"
    extra_hosts:
      - "host.docker.internal:host-gateway"
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
EOF
```

### 2. 启动服务

```bash
docker-compose up -d
```

### 3. 验证

```bash
# 查看日志
docker-compose logs -f

# 测试服务
curl -H "Host: es.prod.code27.cn" http://localhost/
```

## 📝 配置说明

### 必填参数

- `UPSTREAM_HOST`: 上游服务器地址（IP:端口）
  - 示例：`47.90.201.196:443`（443表示HTTPS）

### 可选参数

- `LISTEN_PORT`: 监听端口（默认：80）
- `HTTP_PROXY`: HTTP代理地址（中国环境建议配置）
- `HTTPS_PROXY`: HTTPS代理地址（中国环境建议配置）
- `SOCKS_PROXY`: SOCKS5代理地址

### 代理配置说明

**访问宿主机代理：**
```yaml
HTTP_PROXY: "http://host.docker.internal:7897"
HTTPS_PROXY: "http://host.docker.internal:7897"
```

**不需要代理（海外环境）：**
```yaml
# 不设置代理环境变量即可
UPSTREAM_HOST: "47.90.201.196:443"
LISTEN_PORT: "80"
```

## 🔧 常用命令

```bash
# 启动
docker-compose up -d

# 停止
docker-compose down

# 重启
docker-compose restart

# 查看日志
docker-compose logs -f

# 查看状态
docker-compose ps

# 更新镜像
docker-compose pull
docker-compose up -d
```

## ✅ 测试验证

```bash
# 基础测试
curl -H "Host: es.prod.code27.cn" http://localhost/ -I

# 期望返回：HTTP/1.1 302 Found

# 完整测试
curl -L -H "Host: es.prod.code27.cn" http://localhost/ | grep title

# 期望返回：<title>Elastic</title>
```

## 📊 工作原理

```
HTTP请求 → simple-proxy容器 → 宿主机代理 → 上游HTTPS服务器
         (保持Host头)      (可选)    (47.90.201.196:443)
```

## 💡 使用场景

1. **绕过 DNS**：直接访问 IP，但需要特定的 Host 头
2. **通过代理**：上游服务器需要通过代理访问
3. **统一入口**：多个域名指向同一个 IP

## 📚 详细文档

- [部署说明.md](./部署说明.md) - 完整部署文档
- [DOCKER使用说明.md](./DOCKER使用说明.md) - Docker详细说明
- [使用指南.md](./使用指南.md) - 完整使用指南

---

**提示**: 如需在 80 端口运行，确保没有其他服务占用该端口。



