#!/bin/bash

###############################################################################
# LabEquip-BMS 本地开发启动脚本
# 适用于开发环境快速启动
###############################################################################

# 颜色输出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

print_message() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# 检查 MySQL 是否运行
check_mysql() {
    print_message "检查 MySQL 服务..."
    if ! pgrep -x mysqld > /dev/null; then
        print_warning "MySQL 服务未运行,正在启动..."
        # macOS
        if [[ "$OSTYPE" == "darwin"* ]]; then
            brew services start mysql
        # Linux
        else
            sudo systemctl start mysql
        fi
        sleep 3
    fi
    print_message "MySQL 服务运行正常 ✓"
}

# 检查后端依赖
check_backend_deps() {
    print_message "检查后端依赖..."
    if [ ! -d "vendor" ]; then
        print_message "首次运行,下载依赖..."
        go mod download
    fi
}

# 检查前端依赖
check_frontend_deps() {
    print_message "检查前端依赖..."
    if [ ! -d "frontend/node_modules" ]; then
        print_message "首次运行,安装前端依赖(需要几分钟)..."
        cd frontend
        npm install
        cd ..
    fi
}

# 启动后端
start_backend() {
    print_message "启动后端服务..."
    cd server/api
    go run main.go &
    BACKEND_PID=$!
    cd ../..
    print_message "后端服务已启动 (PID: $BACKEND_PID)"
}

# 启动前端
start_frontend() {
    print_message "启动前端服务..."
    cd frontend
    npm start &
    FRONTEND_PID=$!
    cd ..
    print_message "前端服务已启动 (PID: $FRONTEND_PID)"
}

# 主函数
main() {
    clear
    echo ""
    echo "========================================="
    echo "   LabEquip-BMS 开发环境启动"
    echo "========================================="
    echo ""
    
    check_mysql
    check_backend_deps
    check_frontend_deps
    
    echo ""
    print_message "正在启动服务..."
    echo ""
    
    # 启动后端
    print_message "1/2 启动后端..."
    start_backend
    sleep 5
    
    # 启动前端
    print_message "2/2 启动前端..."
    start_frontend
    
    echo ""
    echo "========================================="
    echo "   服务启动完成！"
    echo "========================================="
    echo ""
    echo "访问地址: http://localhost:3000"
    echo ""
    echo "后端 PID: $BACKEND_PID"
    echo "前端 PID: $FRONTEND_PID"
    echo ""
    echo "停止服务:"
    echo "  kill $BACKEND_PID"
    echo "  kill $FRONTEND_PID"
    echo ""
    echo "或者按 Ctrl+C 停止当前脚本"
    echo "========================================="
    echo ""
    
    # 等待用户中断
    wait
}

# 捕获 Ctrl+C
trap "echo ''; print_message '正在停止服务...'; kill $BACKEND_PID $FRONTEND_PID 2>/dev/null; exit" INT

# 执行主函数
main
