# API 对接说明

## 后端响应格式规范

### 成功响应
```json
{
  "retcode": 0,
  "message": "成功",
  "data": {
    // 具体业务数据
  }
}
```

### 错误响应
```json
{
  "retcode": 非0的错误码,
  "message": "错误描述信息",
  "data": null
}
```

## 前端统一错误处理机制

### 框架层自动处理

所有 API 请求的错误都会在 `src/utils/request.js` 中统一处理：

1. **自动检查 retcode**
   - `retcode === 0`: 成功，返回完整响应数据
   - `retcode !== 0`: 错误，自动弹窗显示 message 字段内容

2. **自动弹窗提示**
   - 使用 Ant Design 的 `message.error()` 显示错误信息
   - 无需在每个页面组件中手动处理错误提示

3. **特殊错误码处理**
   - `retcode === 401` 或 `10001`: 未登录/登录过期，自动跳转到登录页
   - `retcode === 403` 或 `10003`: 权限不足，自动跳转到首页

### 业务代码写法

#### ✅ 推荐写法（简洁）

```javascript
// 只处理成功情况，错误会自动提示
const handleSubmit = async (values) => {
  setLoading(true);
  try {
    const res = await someApi(values);
    
    // 只需要检查成功情况
    if (res && res.retcode === 0) {
      message.success('操作成功');
      // 处理成功逻辑
    }
  } catch (error) {
    // 错误已在框架层显示，这里只需记录日志
    console.error('操作失败:', error);
  } finally {
    setLoading(false);
  }
};
```

#### ❌ 不推荐写法（重复处理）

```javascript
// 不要这样写，会导致错误提示重复显示
const handleSubmit = async (values) => {
  try {
    const res = await someApi(values);
    
    if (res.retcode !== 0) {
      message.error(res.message); // ❌ 框架已经处理了，这里不需要
      return;
    }
    
    // 成功逻辑
  } catch (error) {
    message.error('操作失败'); // ❌ 框架已经处理了，这里不需要
  }
};
```

## 常见错误码

| retcode | 含义 | 前端处理 |
|---------|------|----------|
| 0 | 成功 | 正常返回数据 |
| 400 | 参数错误 | 弹窗提示 |
| 401 | 未登录/登录过期 | 跳转登录页 |
| 403 | 权限不足 | 跳转首页 |
| 404 | 资源不存在 | 弹窗提示 |
| 500 | 服务器错误 | 弹窗提示 |
| 10001 | 未授权 | 跳转登录页 |
| 10003 | 权限不足 | 跳转首页 |

## HTTP 状态码处理

除了业务 retcode，框架也会处理 HTTP 状态码错误：

| HTTP Status | 处理方式 |
|-------------|----------|
| 401 | 清除登录信息，跳转登录页 |
| 403 | 提示权限不足 |
| 404 | 提示资源不存在 |
| 500 | 提示服务器错误 |
| 网络错误 | 提示检查网络连接 |

## API 调用示例

### 登录 API

**请求:**
```javascript
await login({
  email: 'user@example.com',
  password: '123456'
});
```

**成功响应:**
```json
{
  "retcode": 0,
  "message": "登录成功",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user_info": {
      "email": "user@example.com",
      "name": "张三",
      "role": 4,
      "phone": "13800138000"
    }
  }
}
```

**错误响应:**
```json
{
  "retcode": 10001,
  "message": "用户名或密码错误",
  "data": null
}
```

### 查询库存 API

**请求:**
```javascript
await searchInventory({
  lab_code: 'LAB001',
  equip_id: 'EQUIP001'
});
```

**成功响应:**
```json
{
  "retcode": 0,
  "message": "success",
  "data": {
    "total": 10,
    "list": [
      {
        "equip_id": "EQUIP001",
        "lab_code": "LAB001",
        "total_qty": 100,
        "available_qty": 50
      }
    ]
  }
}
```

## 调试技巧

### 查看请求日志

打开浏览器控制台，可以看到：

```
发送请求: POST /api/apps/basic/user/user_login {...}
收到响应: /api/apps/basic/user/user_login 200 {...}
```

### 查看错误日志

如果 retcode !== 0，会看到：

```
业务错误: {
  url: '/api/apps/basic/user/user_login',
  retcode: 10001,
  message: '用户名或密码错误',
  data: {...}
}
```

## 后端开发注意事项

### 1. 统一返回格式

**所有接口必须返回包含 retcode 的 JSON 响应：**

```go
// 成功响应
{
    "retcode": 0,
    "message": "success",
    "data": actualData
}

// 错误响应
{
    "retcode": errorCode,
    "message": "错误描述",
    "data": nil
}
```

### 2. 设置正确的 Content-Type

```go
w.Header().Set("Content-Type", "application/json")
```

### 3. CORS 配置

```go
w.Header().Set("Access-Control-Allow-Origin", "*")
w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-Email")
```

### 4. 错误码规范

建议使用统一的错误码体系：
- 1xxxx: 用户相关错误
- 2xxxx: 设备相关错误
- 3xxxx: 库存相关错误
- 4xxxx: 借记相关错误
- 5xxxx: 系统错误

## 测试 API

使用 curl 测试：

```bash
# 成功情况
curl -X POST http://localhost:8080/apps/basic/user/user_login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"123456"}'

# 期望返回
{
  "retcode": 0,
  "message": "success",
  "data": {
    "token": "xxx",
    "user_info": {...}
  }
}

# 错误情况
{
  "retcode": 10001,
  "message": "用户名或密码错误",
  "data": null
}
```

## 前端适配清单

✅ 已完成：
- [x] 请求拦截器：自动添加 Token
- [x] 响应拦截器：统一检查 retcode
- [x] 错误自动弹窗：使用 message.error()
- [x] 登录过期处理：自动跳转登录页
- [x] 权限不足处理：自动跳转首页
- [x] 网络错误提示：友好的错误信息
- [x] 日志输出：方便调试

## 迁移指南

如果要迁移现有代码到新的错误处理机制：

### 1. 删除业务代码中的错误提示

```javascript
// 旧代码
try {
  const res = await someApi();
  if (res.retcode !== 0) {
    message.error(res.message); // ❌ 删除这行
    return;
  }
} catch (error) {
  message.error('操作失败'); // ❌ 删除这行
}

// 新代码
try {
  const res = await someApi();
  if (res && res.retcode === 0) {
    // 只处理成功情况
  }
} catch (error) {
  console.error('操作失败:', error); // ✅ 只记录日志
}
```

### 2. 检查 retcode 而不是 code

```javascript
// 旧代码
if (res.code === 200) { // ❌

// 新代码
if (res.retcode === 0) { // ✅
```

### 3. 从 res.data 获取业务数据

```javascript
const res = await someApi();
if (res && res.retcode === 0) {
  const actualData = res.data; // ✅ 业务数据在 data 字段
}
```

