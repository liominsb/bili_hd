# ==========================================
# 第一阶段：构建阶段 (Builder)
# ==========================================
FROM golang:alpine AS builder

# 1. 切换工作目录
WORKDIR /app

# 2. 设置 Go 环境变量：开启模块支持、配置国内快速代理
ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct

# 3. 利用 Docker 层缓存：先只复制依赖清单并下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 4. 复制代码全貌
COPY . .

# 5. 静态编译 Go 程序
# CGO_ENABLED=0：关闭 CGO，生成纯静态二进制，不依赖任何系统的动态链接库（极为关键！）
# GOOS=linux：跨平台交叉编译为 Linux 格式
# -ldflags="-s -w"：去除调试符号和 DWARF 信息，让可执行文件体积再缩减 30%~40%
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server .

# ==========================================
# 第二阶段：极简运行环境 (Runner)
# ==========================================
FROM alpine:latest

# 1. 安装基础工具：ca 证书（支持 HTTPS）与时区数据（设置上海时区，保证日志和数据库时间准确）
RUN apk add --no-cache ca-certificates tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

WORKDIR /app

# 2. 从上一阶段 (builder) 仅把编译好的二进制文件与配置文件复制过来
COPY --from=builder /app/server .
COPY --from=builder /app/config ./config

# 3. 创建上传目录（以防业务代码报错）
RUN mkdir -p uploads

# 4. 声明容器监听端口（与 config.yml 中的 3000 端口对应）
EXPOSE 3000

# 5. 容器启动入口
CMD ["./server"]