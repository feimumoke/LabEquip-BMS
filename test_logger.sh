#!/bin/bash

###############################################################################
# 日志系统测试脚本
###############################################################################

echo "========================================="
echo "   LabEquip-BMS 日志系统测试"
echo "========================================="
echo ""

# 检查 logs 目录
if [ ! -d "logs" ]; then
    echo "❌ logs 目录不存在，正在创建..."
    mkdir -p logs
    chmod 755 logs
fi

echo "✅ logs 目录存在"
echo ""

# 检查配置文件
if [ ! -f "server/_config/conf.yaml" ]; then
    echo "❌ 配置文件不存在: server/_config/conf.yaml"
    exit 1
fi

echo "✅ 配置文件存在"
echo ""

# 检查日志配置
if grep -q "^log:" server/_config/conf.yaml; then
    echo "✅ 日志配置已添加"
    echo ""
    echo "日志配置内容:"
    grep -A 3 "^log:" server/_config/conf.yaml
else
    echo "❌ 日志配置未添加到 conf.yaml"
    exit 1
fi

echo ""
echo "========================================="
echo "   测试完成！"
echo "========================================="
echo ""
echo "下一步:"
echo "1. 启动后端服务: cd server/api && go run main.go"
echo "2. 查看日志文件: tail -f logs/bms-$(date +%Y-%m-%d).log"
echo ""
