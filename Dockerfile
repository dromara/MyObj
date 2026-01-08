# 构建阶段
FROM golang:1.25-alpine AS builder

# 设置工作目录
WORKDIR /build

# 安装必要的构建工具
RUN apk add --no-cache git gcc g++ musl-dev

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o myobj ./src/cmd/server/main.go

# 运行阶段
FROM alpine:latest

# 设置工作目录
WORKDIR /app

# 安装运行时依赖
RUN apk --no-cache add ca-certificates tzdata

# 设置时区为上海
ENV TZ=Asia/Shanghai

# 创建必要的目录
RUN mkdir -p /app/logs \
    /app/libs \
    /app/obj_data \
    /app/obj_temp \
    /app/webview/dist

# 从构建阶段复制可执行文件
COPY --from=builder /build/myobj .

# 复制前端静态文件
COPY --from=builder /build/webview/dist ./webview/dist

# 复制模板文件
COPY --from=builder /build/templates ./templates

# 复制 docs 目录（Swagger 文档）
COPY --from=builder /build/docs ./docs

# 暴露端口
# 8080: HTTP服务端口
# 8081: WebDAV服务端口
EXPOSE 8080 8081

# 设置挂载点
VOLUME ["/app/config.toml", "/app/logs", "/app/libs", "/app/obj_data", "/app/obj_temp"]

# 启动应用
CMD ["./myobj"]
