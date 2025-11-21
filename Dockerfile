# 多阶段构建 - 编译阶段
FROM golang:1.21-alpine AS builder

WORKDIR /build

# 安装依赖
RUN apk add --no-cache git

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 编译
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.Version=${VERSION} -X main.BuildTime=$(date -u '+%Y-%m-%dT%H:%M:%SZ')" \
    -o /build/app ./cmd/api

# 多阶段构建 - 运行阶段
FROM alpine:3.18

# 安装运行时依赖
RUN apk add --no-cache ca-certificates tzdata curl

# 创建非 root 用户
RUN addgroup -g 1000 app && adduser -D -u 1000 -G app app

WORKDIR /app

# 从 builder 复制二进制文件
COPY --from=builder /build/app /app/app

# 复制配置文件
COPY config/ /app/config/
COPY migrations/ /app/migrations/

# 改变所有权
RUN chown -R app:app /app

# 切换到非 root 用户
USER app

# 暴露端口
EXPOSE 8080 9090

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# 启动应用
CMD ["/app/app"]

