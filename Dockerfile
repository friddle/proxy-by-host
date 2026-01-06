# 构建阶段
FROM docker.linkos.org/library/golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖（使用中国镜像加速）
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod download

# 复制源代码
COPY main.go ./

# 编译
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o simple-proxy main.go

# 运行阶段
FROM docker.linkos.org/library/alpine:latest

# 安装 CA 证书（用于 HTTPS）
RUN apk --no-cache add ca-certificates tzdata

# 设置时区为上海
ENV TZ=Asia/Shanghai

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/simple-proxy .

# 暴露端口
EXPOSE 80

# 运行
CMD ["./simple-proxy"]

