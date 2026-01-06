.PHONY: build run clean docker-build docker-run test

# 默认目标
all: build

# 编译
build:
	mkdir -p output
	go build -o output/reserver-proxy main.go

# 运行
run:
	go run main.go --ssl=generate

# 清理
clean:
	rm -rf output

# 构建 Docker 镜像
docker-build:
	docker build -t reserver-proxy:latest .

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
	mkdir -p output
	GOOS=linux GOARCH=amd64 go build -o output/reserver-proxy-linux-amd64 main.go
	GOOS=linux GOARCH=arm64 go build -o output/reserver-proxy-linux-arm64 main.go
	GOOS=darwin GOARCH=amd64 go build -o output/reserver-proxy-darwin-amd64 main.go
	GOOS=darwin GOARCH=arm64 go build -o output/reserver-proxy-darwin-arm64 main.go
	GOOS=windows GOARCH=amd64 go build -o output/reserver-proxy-windows-amd64.exe main.go

# 测试
test:
	./output/reserver-proxy --ssl=generate --http-port=8080 --https-port=8443



