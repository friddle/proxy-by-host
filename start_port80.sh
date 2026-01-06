#!/bin/bash

# 启动脚本 - 监听80端口
# 需要 sudo 权限

echo "=== Simple Proxy 启动脚本（80端口）==="
echo ""
echo "配置："
echo "  监听端口: 80"
echo "  上游服务器: 47.252.16.154:443"
echo "  HTTP代理: http://localhost:7897"
echo "  HTTPS代理: http://localhost:7897"
echo ""

# 停止可能存在的旧进程
sudo pkill -9 simple-proxy 2>/dev/null

# 启动服务
echo "正在启动服务（需要sudo密码）..."
sudo UPSTREAM_HOST=47.252.16.154:443 \
     LISTEN_PORT=80 \
     HTTP_PROXY=http://localhost:7897 \
     HTTPS_PROXY=http://localhost:7897 \
     /Users/friddle/Project/simple_proxy/simple-proxy &

sleep 2

# 检查服务状态
if lsof -i :80 | grep simple-proxy > /dev/null 2>&1; then
    echo "✅ 服务启动成功！监听端口 80"
    echo ""
    echo "测试命令："
    echo "  curl -H 'Host: es.prod.code27.cn' http://localhost/"
    echo ""
    echo "或者配置 /etc/hosts 后直接访问："
    echo "  echo '127.0.0.1 es.prod.code27.cn' | sudo tee -a /etc/hosts"
    echo "  curl http://es.prod.code27.cn/"
else
    echo "❌ 服务启动失败，请检查日志"
fi



