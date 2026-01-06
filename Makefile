.PHONY: build run clean docker-build docker-run test

# 默认目标
all: build

# 编译
build:
	go build -o simple-proxy main.go

# 运行
run:
	go run main.go

# 清理
clean:
	rm -f simple-proxy simple-proxy-*

# 构建 Docker 镜像
docker-build:
	docker build -t simple-proxy:latest .

# 运行 Docker 容器
docker-run:
	docker-compose up -d

# 停止 Docker 容器
docker-stop:
	docker-compose down

# 查看日志
logs:
	docker-compose logs -f

# 跨平台编译
build-all:
	GOOS=linux GOARCH=amd64 go build -o simple-proxy-linux-amd64 main.go
	GOOS=darwin GOARCH=amd64 go build -o simple-proxy-darwin-amd64 main.go
	GOOS=darwin GOARCH=arm64 go build -o simple-proxy-darwin-arm64 main.go
	GOOS=windows GOARCH=amd64 go build -o simple-proxy-windows-amd64.exe main.go

# 测试
test:
	@echo "启动测试服务器..."
	@export UPSTREAM_HOST=47.252.16.154 && \
	export LISTEN_PORT=8080 && \
	echo "请在另一个终端运行: curl -H 'Host: webmail.prod.code27.cn' http://localhost:8080/"



