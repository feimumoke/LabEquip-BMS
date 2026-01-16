#!/bin/bash

###############################################################################
# LabEquip-BMS 一键部署脚本
# 适用于 Linux/macOS 系统
###############################################################################

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_message() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查命令是否存在
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# 检查环境
check_environment() {
    print_message "正在检查部署环境..."
    
    # 检查 Docker
    if ! command_exists docker; then
        print_error "Docker 未安装，请先安装 Docker"
        exit 1
    fi
    
    # 检查 Docker Compose
    if ! command_exists docker-compose; then
        print_error "Docker Compose 未安装，请先安装 Docker Compose"
        exit 1
    fi
    
    print_message "环境检查通过 ✓"
}

# 创建必要的目录
create_directories() {
    print_message "创建必要的目录..."
    
    mkdir -p uploads
    chmod 755 uploads
    
    mkdir -p logs
    chmod 755 logs
    
    print_message "目录创建完成 ✓"
}

# 检查配置文件
check_config() {
    print_message "检查配置文件..."
    
    if [ ! -f "server/_config/conf.yaml" ]; then
        print_error "配置文件不存在: server/_config/conf.yaml"
        exit 1
    fi
    
    if [ ! -f "database_schema.sql" ]; then
        print_error "数据库初始化脚本不存在: database_schema.sql"
        exit 1
    fi
    
    print_message "配置文件检查完成 ✓"
}

# 构建并启动服务
deploy_services() {
    print_message "开始部署服务..."
    
    # 停止旧的容器
    print_message "停止旧的容器..."
    docker-compose down
    
    # 构建镜像
    print_message "构建 Docker 镜像..."
    docker-compose build --no-cache
    
    # 启动服务
    print_message "启动服务..."
    docker-compose up -d
    
    # 等待服务启动
    print_message "等待服务启动..."
    sleep 15
    
    # 检查服务状态
    print_message "检查服务状态..."
    docker-compose ps
}

# 显示访问信息
show_access_info() {
    echo ""
    echo "========================================="
    echo "   LabEquip-BMS 部署成功！"
    echo "========================================="
    echo ""
    echo "访问地址："
    echo "  前端: http://localhost"
    echo "  后端 API: http://localhost:8080"
    echo ""
    echo "查看日志："
    echo "  docker-compose logs -f"
    echo ""
    echo "停止服务："
    echo "  docker-compose down"
    echo ""
    echo "重启服务："
    echo "  docker-compose restart"
    echo ""
    echo "========================================="
}

# 主函数
main() {
    echo ""
    echo "========================================="
    echo "   LabEquip-BMS 一键部署脚本"
    echo "========================================="
    echo ""
    
    check_environment
    create_directories
    check_config
    deploy_services
    show_access_info
}

# 执行主函数
main
