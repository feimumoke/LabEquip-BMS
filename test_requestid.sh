#!/bin/bash

# RequestID 功能测试脚本

echo "=========================================="
echo "RequestID 追踪功能测试"
echo "=========================================="
echo ""

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查服务是否启动
echo -n "检查服务状态... "
if curl -s http://localhost:8080/api/ping > /dev/null 2>&1; then
    echo -e "${GREEN}✓ 服务已启动${NC}"
else
    echo -e "${RED}✗ 服务未启动${NC}"
    echo "请先启动服务: cd server/api && go run main.go"
    exit 1
fi

echo ""

# 测试 1: 不传递 RequestID（自动生成）
echo "测试 1: 自动生成 RequestID"
echo "----------------------------"
response=$(curl -s -i http://localhost:8080/api/enum/list 2>&1)
request_id=$(echo "$response" | grep -i "X-Request-ID:" | awk '{print $2}' | tr -d '\r')

if [ -n "$request_id" ]; then
    echo -e "${GREEN}✓ 服务器自动生成了 RequestID${NC}"
    echo "  RequestID: $request_id"
    
    # 检查日志中是否包含此 RequestID
    log_file="logs/bms-$(date +%Y-%m-%d).log"
    if [ -f "$log_file" ]; then
        log_count=$(grep -c "$request_id" "$log_file" 2>/dev/null || echo "0")
        if [ "$log_count" -gt 0 ]; then
            echo -e "${GREEN}✓ 日志中找到 $log_count 条相关记录${NC}"
            echo ""
            echo "相关日志："
            grep "$request_id" "$log_file" | tail -3
        else
            echo -e "${YELLOW}⚠ 日志文件中未找到相关记录（可能还未写入）${NC}"
        fi
    else
        echo -e "${YELLOW}⚠ 日志文件不存在: $log_file${NC}"
    fi
else
    echo -e "${RED}✗ 未找到 X-Request-ID 响应头${NC}"
fi

echo ""
echo ""

# 测试 2: 手动传递 RequestID
echo "测试 2: 客户端传递自定义 RequestID"
echo "----------------------------"
custom_request_id="test-custom-$(date +%s)-12345"
echo "自定义 RequestID: $custom_request_id"

response=$(curl -s -i -H "X-Request-ID: $custom_request_id" http://localhost:8080/api/enum/list 2>&1)
returned_id=$(echo "$response" | grep -i "X-Request-ID:" | awk '{print $2}' | tr -d '\r')

if [ "$returned_id" = "$custom_request_id" ]; then
    echo -e "${GREEN}✓ 服务器返回了相同的 RequestID${NC}"
    
    # 等待日志写入
    sleep 1
    
    # 检查日志
    log_file="logs/bms-$(date +%Y-%m-%d).log"
    if [ -f "$log_file" ]; then
        log_count=$(grep -c "$custom_request_id" "$log_file" 2>/dev/null || echo "0")
        if [ "$log_count" -gt 0 ]; then
            echo -e "${GREEN}✓ 日志中找到 $log_count 条相关记录${NC}"
            echo ""
            echo "相关日志："
            grep "$custom_request_id" "$log_file" | tail -3
        else
            echo -e "${YELLOW}⚠ 日志文件中未找到相关记录（可能还未写入）${NC}"
        fi
    fi
else
    echo -e "${RED}✗ 返回的 RequestID 不匹配${NC}"
    echo "  期望: $custom_request_id"
    echo "  实际: $returned_id"
fi

echo ""
echo ""

# 测试 3: 连续发送多个请求
echo "测试 3: 连续请求的 RequestID 唯一性"
echo "----------------------------"
request_ids=()
for i in {1..5}; do
    response=$(curl -s -i http://localhost:8080/api/enum/list 2>&1)
    request_id=$(echo "$response" | grep -i "X-Request-ID:" | awk '{print $2}' | tr -d '\r')
    request_ids+=("$request_id")
    echo "  请求 $i: $request_id"
done

# 检查唯一性
unique_count=$(printf '%s\n' "${request_ids[@]}" | sort -u | wc -l | tr -d ' ')
total_count=${#request_ids[@]}

if [ "$unique_count" -eq "$total_count" ]; then
    echo -e "${GREEN}✓ 所有 RequestID 都是唯一的 ($unique_count/$total_count)${NC}"
else
    echo -e "${RED}✗ 发现重复的 RequestID ($unique_count/$total_count)${NC}"
fi

echo ""
echo ""

# 测试总结
echo "=========================================="
echo "测试完成"
echo "=========================================="
echo ""
echo "手动验证建议："
echo "1. 查看实时日志:"
echo "   tail -f logs/bms-\$(date +%Y-%m-%d).log"
echo ""
echo "2. 根据 RequestID 追踪请求:"
echo "   grep \"$custom_request_id\" logs/bms-\$(date +%Y-%m-%d).log"
echo ""
echo "3. 查看所有包含 RequestID 的日志:"
echo "   grep \"\\[RequestID:\" logs/bms-\$(date +%Y-%m-%d).log | tail -20"
echo ""
