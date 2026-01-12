#!/bin/bash

# 测试登录态校验脚本

BASE_URL="http://localhost:8080/api"

echo "========================================="
echo "测试 1: 登录获取 token"
echo "========================================="

LOGIN_RESPONSE=$(curl -s -X POST "${BASE_URL}/apps/basic/user/user_login" \
  -H "Content-Type: application/json" \
  -d '{
    "client_type": 1,
    "code": "admin@example.com",
    "passwd": "admin123"
  }')

echo "登录响应:"
echo "$LOGIN_RESPONSE" | jq '.'

# 提取 token
TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.token')
USER_EMAIL=$(echo "$LOGIN_RESPONSE" | jq -r '.data.user_info.email')

if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
    echo "❌ 登录失败，无法获取 token"
    exit 1
fi

echo ""
echo "✅ 登录成功"
echo "Token: $TOKEN"
echo "Email: $USER_EMAIL"

echo ""
echo "========================================="
echo "测试 2: 使用 token 访问受保护的 API"
echo "========================================="

PROTECTED_RESPONSE=$(curl -s -X POST "${BASE_URL}/apps/bms/inventory/search_equip_inv" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-User-Email: ${USER_EMAIL}" \
  -d '{}')

echo "受保护 API 响应:"
echo "$PROTECTED_RESPONSE" | jq '.'

RETCODE=$(echo "$PROTECTED_RESPONSE" | jq -r '.retcode')
if [ "$RETCODE" == "0" ]; then
    echo "✅ 使用 token 访问成功"
else
    echo "❌ 访问失败，retcode: $RETCODE"
fi

echo ""
echo "========================================="
echo "测试 3: 不带 token 访问受保护的 API"
echo "========================================="

UNAUTH_RESPONSE=$(curl -s -X POST "${BASE_URL}/apps/bms/inventory/search_equip_inv" \
  -H "Content-Type: application/json" \
  -d '{}')

echo "未授权访问响应:"
echo "$UNAUTH_RESPONSE" | jq '.'

RETCODE=$(echo "$UNAUTH_RESPONSE" | jq -r '.retcode')
if [ "$RETCODE" != "0" ]; then
    echo "✅ 正确拒绝未授权访问"
else
    echo "❌ 安全问题：未授权访问成功了"
fi

echo ""
echo "========================================="
echo "测试 4: 使用无效 token 访问"
echo "========================================="

INVALID_RESPONSE=$(curl -s -X POST "${BASE_URL}/apps/bms/inventory/search_equip_inv" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer invalid_token_12345" \
  -H "X-User-Email: ${USER_EMAIL}" \
  -d '{}')

echo "无效 token 响应:"
echo "$INVALID_RESPONSE" | jq '.'

RETCODE=$(echo "$INVALID_RESPONSE" | jq -r '.retcode')
if [ "$RETCODE" != "0" ]; then
    echo "✅ 正确拒绝无效 token"
else
    echo "❌ 安全问题：无效 token 访问成功了"
fi

echo ""
echo "========================================="
echo "测试 5: 访问白名单路径（不需要 token）"
echo "========================================="

WHITELIST_RESPONSE=$(curl -s -X POST "${BASE_URL}/apps/common/enums" \
  -H "Content-Type: application/json" \
  -d '{}')

echo "白名单路径响应:"
echo "$WHITELIST_RESPONSE" | jq '.'

RETCODE=$(echo "$WHITELIST_RESPONSE" | jq -r '.retcode')
if [ "$RETCODE" == "0" ]; then
    echo "✅ 白名单路径访问成功"
else
    echo "❌ 白名单路径访问失败"
fi

echo ""
echo "========================================="
echo "测试完成！"
echo "========================================="
