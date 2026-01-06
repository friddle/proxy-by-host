#!/bin/bash

# Docker 镜像构建和推送脚本

set -e

# 配置
IMAGE_NAME="registrylan.service.code27.cn/third/simple_proxy"
VERSION=${1:-latest}

echo "==================================="
echo "Simple Proxy Docker 镜像构建和推送"
echo "==================================="
echo ""
echo "镜像名称: $IMAGE_NAME"
echo "版本标签: $VERSION"
echo ""

# 构建镜像
echo ">>> 步骤 1: 构建 Docker 镜像 (平台: linux/amd64)..."
docker build --platform linux/amd64 -t ${IMAGE_NAME}:${VERSION} .

if [ $? -eq 0 ]; then
    echo "✅ 镜像构建成功!"
else
    echo "❌ 镜像构建失败!"
    exit 1
fi

# 如果不是 latest，也打上 latest 标签
if [ "$VERSION" != "latest" ]; then
    echo ""
    echo ">>> 步骤 2: 打上 latest 标签..."
    docker tag ${IMAGE_NAME}:${VERSION} ${IMAGE_NAME}:latest
fi

# 推送镜像
echo ""
echo ">>> 步骤 3: 推送镜像到仓库..."
docker push ${IMAGE_NAME}:${VERSION}

if [ $? -eq 0 ]; then
    echo "✅ ${IMAGE_NAME}:${VERSION} 推送成功!"
else
    echo "❌ 镜像推送失败!"
    exit 1
fi

if [ "$VERSION" != "latest" ]; then
    echo ""
    echo ">>> 步骤 4: 推送 latest 标签..."
    docker push ${IMAGE_NAME}:latest
    
    if [ $? -eq 0 ]; then
        echo "✅ ${IMAGE_NAME}:latest 推送成功!"
    else
        echo "❌ latest 标签推送失败!"
        exit 1
    fi
fi

echo ""
echo "==================================="
echo "✅ 全部完成!"
echo "==================================="
echo ""
echo "镜像地址:"
echo "  ${IMAGE_NAME}:${VERSION}"
if [ "$VERSION" != "latest" ]; then
    echo "  ${IMAGE_NAME}:latest"
fi
echo ""
echo "使用方式:"
echo "  docker pull ${IMAGE_NAME}:${VERSION}"
echo "  docker-compose up -d"
echo ""


