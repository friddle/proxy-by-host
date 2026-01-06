#!/bin/bash

# 简单代理启动脚本示例

echo "=== Simple Proxy 启动脚本 ==="
echo ""

# 配置参数
export UPSTREAM_HOST=${UPSTREAM_HOST:-47.252.16.154}
export LISTEN_PORT=${LISTEN_PORT:-8080}

# 代理配置（根据实际情况选择一种）
# 示例 1: 使用 HTTP 代理（如 Clash、V2rayN 等）
# export HTTP_PROXY=http://127.0.0.1:7890
# export HTTPS_PROXY=http://127.0.0.1:7890

# 示例 2: 使用 SOCKS5 代理
# export SOCKS_PROXY=socks5://127.0.0.1:1080

# 示例 3: 使用需要认证的代理
# export HTTP_PROXY=http://username:password@127.0.0.1:7890
# export SOCKS_PROXY=socks5://username:password@127.0.0.1:1080

echo "配置信息："
echo "  上游服务器: $UPSTREAM_HOST"
echo "  监听端口: $LISTEN_PORT"
echo "  HTTP 代理: ${HTTP_PROXY:-未设置}"
echo "  HTTPS 代理: ${HTTPS_PROXY:-未设置}"
echo "  SOCKS 代理: ${SOCKS_PROXY:-未设置}"
echo ""

# 检查是否已编译
if [ ! -f "./simple-proxy" ]; then
    echo "未找到编译好的程序，正在编译..."
    go build -o simple-proxy main.go
    if [ $? -ne 0 ]; then
        echo "编译失败！"
        exit 1
    fi
    echo "编译成功！"
fi

echo ""
echo "正在启动代理服务..."
echo "按 Ctrl+C 停止服务"
echo ""
echo "测试命令："
echo "  curl -H 'Host: webmail.prod.code27.cn' http://localhost:$LISTEN_PORT/"
echo ""

./simple-proxy



