# 前端构建阶段
FROM node:24-alpine AS frontend-builder

# 安装固定版本的 pnpm（确保构建一致性）
RUN npm install -g pnpm@latest

RUN pnpm config set registry https://registry.npmmirror.com/

WORKDIR /build

# 复制整个 webview 目录（.dockerignore 会排除 node_modules 和 dist）
COPY webview /build/webview

WORKDIR /build/webview

# 安装前端依赖
RUN pnpm install

# 构建前端（使用生产模式，设置 NODE_ENV）
ENV NODE_ENV=production
RUN pnpm run build:prod || pnpm run build

# Go 后端构建阶段
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

# 从前端构建阶段复制构建好的前端文件
COPY --from=frontend-builder /build/webview/dist ./webview/dist

# 构建应用（支持多架构）
ARG TARGETOS=linux
ARG TARGETARCH
RUN CGO_ENABLED=1 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -a -installsuffix cgo -o myobj ./src/cmd/server/main.go

# 运行阶段
FROM alpine:latest

# 设置工作目录
WORKDIR /app

# 安装运行时依赖（包括 C++ 运行时库，因为使用了 CGO）
RUN apk --no-cache add ca-certificates tzdata libstdc++ libgcc

# 设置时区为上海
ENV TZ=Asia/Shanghai

# 标识运行在 Docker 容器中
ENV MYOBJ_IN_DOCKER=1

# 创建必要的目录
RUN mkdir -p /app/logs \
    /app/libs \
    /app/obj_data \
    /app/obj_temp \
    /app/webview/dist \
    /app/default-libs

# 从构建阶段复制可执行文件
COPY --from=builder /build/myobj .

# 复制前端静态文件
COPY --from=builder /build/webview/dist ./webview/dist

# 复制模板文件
COPY --from=builder /build/templates ./templates

# 复制 docs 目录（Swagger 文档）
COPY --from=builder /build/docs ./docs

# 复制默认数据库文件（用于初始化）
COPY libs/my_obj.db /app/default-libs/my_obj.db

# 复制初始化脚本
COPY docker-entrypoint.sh /app/docker-entrypoint.sh
RUN chmod +x /app/docker-entrypoint.sh

# 暴露端口
# 8080: HTTP服务端口
# 8081: WebDAV服务端口
EXPOSE 8080 8081

# 设置挂载点
VOLUME ["/app/config.toml", "/app/logs", "/app/libs", "/app/obj_data", "/app/obj_temp"]

# 设置入口点（运行初始化脚本后启动应用）
ENTRYPOINT ["/app/docker-entrypoint.sh"]
CMD ["./myobj"]
