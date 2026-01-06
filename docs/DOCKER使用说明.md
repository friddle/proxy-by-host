# Simple Proxy Docker 使用说明

## 🚀 快速开始

### 方式一：使用 docker-compose（推荐）

1. **创建配置文件**

```bash
# 创建 docker-compose.yml
cat > docker-compose.yml << 'EOF'
version: '3.8'

services:
  simple-proxy:
    image: registrylan.service.code27.cn/third/simple_proxy:latest
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
EOF
```

2. **启动服务**

```bash
docker-compose up -d
```

3. **查看日志**

```bash
docker-compose logs -f
```

4. **测试**

```bash
curl -H "Host: es.prod.code27.cn" http://localhost/
```

### 方式二：使用 docker run

```bash
docker run -d \
  --name simple-proxy \
  --restart unless-stopped \
  -p 80:80 \
  -e UPSTREAM_HOST=47.90.201.196:443 \
  -e LISTEN_PORT=80 \
  -e HTTP_PROXY=http://host.docker.internal:7897 \
  -e HTTPS_PROXY=http://host.docker.internal:7897 \
  --add-host host.docker.internal:host-gateway \
  registrylan.service.code27.cn/third/simple_proxy:latest
```

## 📝 环境变量配置

| 变量名 | 说明 | 默认值 | 示例 |
|--------|------|--------|------|
| `UPSTREAM_HOST` | 上游服务器 IP:端口 | `47.90.201.196:443` | `47.90.201.196:443` |
| `LISTEN_PORT` | 容器内监听端口 | `80` | `80` |
| `HTTP_PROXY` | HTTP 代理地址 | - | `http://host.docker.internal:7897` |
| `HTTPS_PROXY` | HTTPS 代理地址 | - | `http://host.docker.internal:7897` |
| `SOCKS_PROXY` | SOCKS5 代理地址 | - | `socks5://host.docker.internal:1080` |

**注意：**
- 容器内访问宿主机服务使用 `host.docker.internal`
- 端口 443 表示上游使用 HTTPS 协议
- 代理地址可选，如果上游服务器可直接访问则不需要

## 🔧 常见配置场景

### 场景 1: 不需要代理（上游可直连）

```yaml
environment:
  UPSTREAM_HOST: "47.90.201.196:443"
  LISTEN_PORT: "80"
```

### 场景 2: 通过宿主机代理访问（中国环境）

```yaml
environment:
  UPSTREAM_HOST: "47.90.201.196:443"
  LISTEN_PORT: "80"
  HTTP_PROXY: "http://host.docker.internal:7897"
  HTTPS_PROXY: "http://host.docker.internal:7897"
extra_hosts:
  - "host.docker.internal:host-gateway"
```

### 场景 3: 使用 SOCKS5 代理

```yaml
environment:
  UPSTREAM_HOST: "47.90.201.196:443"
  LISTEN_PORT: "80"
  SOCKS_PROXY: "socks5://host.docker.internal:1080"
extra_hosts:
  - "host.docker.internal:host-gateway"
```

### 场景 4: 部署多个实例（不同上游）

```yaml
version: '3.8'

services:
  proxy-es:
    image: registrylan.service.code27.cn/third/simple_proxy:latest
    container_name: proxy-es
    ports:
      - "8081:80"
    environment:
      UPSTREAM_HOST: "47.90.201.196:443"
      HTTP_PROXY: "http://host.docker.internal:7897"
      HTTPS_PROXY: "http://host.docker.internal:7897"
    extra_hosts:
      - "host.docker.internal:host-gateway"

  proxy-other:
    image: registrylan.service.code27.cn/third/simple_proxy:latest
    container_name: proxy-other
    ports:
      - "8082:80"
    environment:
      UPSTREAM_HOST: "47.252.16.154:443"
      HTTP_PROXY: "http://host.docker.internal:7897"
      HTTPS_PROXY: "http://host.docker.internal:7897"
    extra_hosts:
      - "host.docker.internal:host-gateway"
```

## 🏗️ 构建镜像

### 构建并推送到私有仓库

```bash
# 给脚本执行权限
chmod +x build-and-push.sh

# 构建并推送 latest 版本
./build-and-push.sh

# 构建并推送指定版本
./build-and-push.sh v1.0.0
```

### 手动构建

```bash
# 构建镜像
docker build -t registrylan.service.code27.cn/third/simple_proxy:latest .

# 推送到仓库
docker push registrylan.service.code27.cn/third/simple_proxy:latest
```

## 📊 监控和管理

### 查看容器状态

```bash
docker ps | grep simple-proxy
```

### 查看实时日志

```bash
docker logs -f simple-proxy
```

### 查看资源使用

```bash
docker stats simple-proxy
```

### 进入容器调试

```bash
docker exec -it simple-proxy sh
```

### 重启服务

```bash
docker-compose restart
# 或
docker restart simple-proxy
```

### 停止并删除

```bash
docker-compose down
# 或
docker stop simple-proxy && docker rm simple-proxy
```

## 🧪 测试验证

### 基本测试

```bash
# 测试服务是否响应
curl -I -H "Host: es.prod.code27.cn" http://localhost/

# 应该返回 302 Found
```

### 完整测试

```bash
# 跟随重定向获取完整页面
curl -L -H "Host: es.prod.code27.cn" http://localhost/ | grep "<title>"

# 应该返回 <title>Elastic</title>
```

### 验证代理工作

查看容器日志，应该看到：

```
2025/10/20 22:30:00 使用 HTTP/HTTPS 代理
2025/10/20 22:30:00 上游服务器: 47.90.201.196:443
2025/10/20 22:30:00 HTTP代理: http://host.docker.internal:7897
2025/10/20 22:30:01 转发请求到: https://47.90.201.196:443/ (Host头: es.prod.code27.cn)
2025/10/20 22:30:02 请求完成: GET / -> 302 (0 bytes)
```

## 🔒 安全建议

1. **不要暴露到公网**
   - 仅在内网或 VPN 中使用
   - 如需公网访问，添加认证层（Nginx + Basic Auth）

2. **使用特定版本**
   ```yaml
   image: registrylan.service.code27.cn/third/simple_proxy:v1.0.0
   ```

3. **限制资源**
   ```yaml
   deploy:
     resources:
       limits:
         cpus: '0.5'
         memory: 256M
   ```

4. **使用只读文件系统**
   ```yaml
   read_only: true
   ```

## 📚 参考配置

完整的生产环境配置示例：

```yaml
version: '3.8'

services:
  simple-proxy:
    image: registrylan.service.code27.cn/third/simple_proxy:latest
    container_name: simple-proxy
    restart: unless-stopped
    
    # 端口映射
    ports:
      - "80:80"
    
    # 环境变量
    environment:
      UPSTREAM_HOST: "47.90.201.196:443"
      LISTEN_PORT: "80"
      HTTP_PROXY: "http://host.docker.internal:7897"
      HTTPS_PROXY: "http://host.docker.internal:7897"
    
    # 主机映射
    extra_hosts:
      - "host.docker.internal:host-gateway"
    
    # 日志配置
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
    
    # 健康检查
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:80"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s
    
    # 资源限制（可选）
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 256M
        reservations:
          cpus: '0.1'
          memory: 64M
```

## ❓ 常见问题

### Q1: 容器无法连接到宿主机代理

**A:** 确保：
1. 使用了 `host.docker.internal` 而不是 `localhost`
2. 添加了 `extra_hosts` 配置
3. 宿主机代理监听了 `0.0.0.0` 而不只是 `127.0.0.1`

### Q2: 健康检查失败

**A:** 需要安装 `wget`，已在 Dockerfile 中配置。如果仍失败，检查 `LISTEN_PORT` 是否正确。

### Q3: 如何查看详细错误

**A:** 
```bash
docker logs simple-proxy --tail 100
```

### Q4: 端口已被占用

**A:** 修改端口映射：
```yaml
ports:
  - "8080:80"  # 宿主机使用 8080，容器内仍是 80
```

## 📞 技术支持

遇到问题查看日志：
```bash
docker-compose logs -f simple-proxy
```

日志会显示详细的请求转发信息和错误。



