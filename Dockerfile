# 多阶段构建 - 后端
FROM golang:1.17-alpine AS backend-builder

# 设置工作目录
WORKDIR /app

# 设置 Go 代理(国内加速)
ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct

# 复制 go mod 文件
COPY go.mod go.sum ./
RUN go mod download

# 复制所有源代码
COPY . .

# 构建后端应用
RUN cd server/api && go build -o /app/bms-api main.go

# 多阶段构建 - 前端
FROM node:14-alpine AS frontend-builder

# 设置工作目录
WORKDIR /app

# 设置 npm 镜像(国内加速)
RUN npm config set registry https://registry.npmmirror.com

# 复制 package 文件
COPY frontend/package*.json ./
RUN npm install

# 复制前端源代码
COPY frontend/ ./

# 构建前端应用
RUN npm run build

# 最终镜像
FROM alpine:latest

# 安装必要的运行时依赖
RUN apk --no-cache add ca-certificates tzdata

# 设置时区为东八区
ENV TZ=Asia/Shanghai

# 创建工作目录
WORKDIR /root/

# 复制后端可执行文件
COPY --from=backend-builder /app/bms-api .

# 复制配置文件
COPY --from=backend-builder /app/server/_config ./server/_config

# 复制前端构建文件
COPY --from=frontend-builder /app/build ./frontend/build

# 创建上传目录
RUN mkdir -p uploads

# 暴露端口
EXPOSE 8080

# 启动应用
CMD ["./bms-api"]
