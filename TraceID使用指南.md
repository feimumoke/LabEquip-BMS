# TraceID 使用指南

## 概述

系统已升级为使用 TraceID（也称为 RequestID）来追踪请求链路。TraceID 是分布式追踪系统中的标准命名。

## 核心函数：GetOrNewTraceID

### 函数签名

```go
// 从 framework/web 包
func GetOrNewTraceID(ctx context.Context) string

// 从 framework/log 包
func GetOrNewTraceID(ctx context.Context) string
```

### 功能说明

- 如果 context 中已存在 TraceID，则返回现有的
- 如果 context 中不存在 TraceID，则自动生成一个新的
- 适用于任何需要 TraceID 的场景

## 使用场景

### 场景 1：HTTP 请求（自动注入）

HTTP 请求通过中间件自动注入 TraceID，**无需手动调用**：

```go
func UserHandler(c *gin.Context) {
    ctx := c.Request.Context()
    
    // TraceID 已自动注入到 ctx 中
    // 直接使用 CtxXXXf 日志函数即可
    log.CtxInfof(ctx, "Processing user request\n")
    
    // 传递给 Service 层
    userService.Process(ctx, data)
}
```

### 场景 2：定时任务（手动获取或创建）

```go
import (
    "context"
    "github.com/feimumoke/labequipbms/framework/log"
)

func CronTask() {
    // 方式 1: 创建新 context 并注入 TraceID
    ctx := context.Background()
    traceID := log.GetOrNewTraceID(ctx)
    ctx = context.WithValue(ctx, "trace_id", traceID)
    
    log.CtxInfof(ctx, "Cron task started\n")
    processCronTask(ctx)
}

// 方式 2: 更简洁的写法
func CronTaskSimple() {
    ctx := context.Background()
    
    // 第一次调用 CtxXXXf 会尝试获取 TraceID
    // 如果没有，会在日志中显示为空
    // 所以建议先手动设置
    traceID := log.GetOrNewTraceID(ctx)
    ctx = context.WithValue(ctx, "trace_id", traceID)
    
    log.CtxInfof(ctx, "Cron task started\n")
}
```

### 场景 3：消息队列消费者

```go
func ConsumeMessage(msg *Message) {
    // 尝试从消息中获取 TraceID
    traceID := msg.Headers["X-Trace-ID"]
    
    ctx := context.Background()
    if traceID == "" {
        // 如果消息中没有 TraceID，生成一个新的
        traceID = log.GetOrNewTraceID(ctx)
    }
    
    // 注入到 context
    ctx = context.WithValue(ctx, "trace_id", traceID)
    
    log.CtxInfof(ctx, "Processing message: %s\n", msg.ID)
    processMessage(ctx, msg)
}
```

### 场景 4：异步任务

```go
func ProcessAsync(ctx context.Context, data string) {
    // 确保异步任务有 TraceID
    traceID := log.GetOrNewTraceID(ctx)
    
    go func() {
        // 创建新 context 并继承 TraceID
        asyncCtx := context.WithValue(context.Background(), "trace_id", traceID)
        
        log.CtxInfof(asyncCtx, "Async task started: %s\n", data)
        // ... 异步处理逻辑
    }()
}
```

### 场景 5：Goroutine 中使用

```go
func SpawnWorkers(ctx context.Context, tasks []Task) {
    // 从父 context 获取或创建 TraceID
    traceID := log.GetOrNewTraceID(ctx)
    
    for _, task := range tasks {
        go func(t Task) {
            // 每个 worker 使用相同的 TraceID
            workerCtx := context.WithValue(context.Background(), "trace_id", traceID)
            
            log.CtxInfof(workerCtx, "Worker processing task: %s\n", t.ID)
            processTask(workerCtx, t)
        }(task)
    }
}
```

### 场景 6：调用外部服务

```go
import "github.com/feimumoke/labequipbms/framework/web"

func CallExternalService(ctx context.Context, url string) (*Response, error) {
    // 获取当前 TraceID
    traceID := web.GetOrNewTraceID(ctx)
    
    // 创建 HTTP 请求
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    
    // 将 TraceID 传递给下游服务
    req.Header.Set("X-Trace-ID", traceID)
    req.Header.Set("X-Request-ID", traceID)  // 兼容
    
    log.CtxInfof(ctx, "Calling external service: %s\n", url)
    
    return http.DefaultClient.Do(req)
}
```

## API 参考

### framework/web 包

```go
// 获取或创建 TraceID（推荐用于 HTTP 场景）
func GetOrNewTraceID(ctx context.Context) string

// 获取 TraceID，不存在返回空字符串
func GetTraceID(ctx context.Context) string

// 从 Gin Context 获取 TraceID
func GetTraceIDFromGin(c *gin.Context) string

// 兼容旧命名
func GetRequestID(ctx context.Context) string
func GetRequestIDFromGin(c *gin.Context) string
```

### framework/log 包

```go
// 获取或创建 TraceID（推荐用于非 HTTP 场景）
func GetOrNewTraceID(ctx context.Context) string

// 带 TraceID 的日志函数
func CtxInfof(ctx context.Context, format string, v ...interface{})
func CtxDebugf(ctx context.Context, format string, v ...interface{})
func CtxWarnf(ctx context.Context, format string, v ...interface{})
func CtxErrorf(ctx context.Context, format string, v ...interface{})
func CtxFatalf(ctx context.Context, format string, v ...interface{})
```

## HTTP 头支持

系统同时支持两种 HTTP 头名称：

- **X-Trace-ID**（推荐，分布式追踪标准）
- **X-Request-ID**（兼容，传统命名）

### 请求示例

```bash
# 使用 X-Trace-ID
curl -H "X-Trace-ID: my-trace-id-12345" http://localhost:8080/api/users

# 使用 X-Request-ID（兼容）
curl -H "X-Request-ID: my-request-id-12345" http://localhost:8080/api/users
```

### 响应示例

服务器会在响应头中同时返回两个头：

```
HTTP/1.1 200 OK
X-Trace-ID: 123e4567-e89b-12d3-a456-426614174000
X-Request-ID: 123e4567-e89b-12d3-a456-426614174000
Content-Type: application/json
```

## 日志输出格式

```
2026/01/16 15:30:45 user_service.go:25: [INFO] [TraceID: 123e4567-e89b-12d3-a456-426614174000] User login successful
2026/01/16 15:30:45 gorm_logger.go:60: [INFO] [TraceID: 123e4567-e89b-12d3-a456-426614174000] [GORM] SELECT * FROM user_tab
```

## 查询日志

```bash
# 根据 TraceID 查询所有相关日志
grep "123e4567-e89b-12d3-a456-426614174000" logs/bms-$(date +%Y-%m-%d).log

# 实时监控所有带 TraceID 的日志
tail -f logs/bms-$(date +%Y-%m-%d).log | grep "\[TraceID:"

# 查找特定时间段的 TraceID
grep "\[TraceID:" logs/bms-2026-01-16.log | grep "15:30"
```

## 最佳实践

### ✅ 推荐做法

```go
// 1. HTTP Handler 中直接使用 context
func Handler(c *gin.Context) {
    ctx := c.Request.Context()  // TraceID 已自动注入
    log.CtxInfof(ctx, "Processing\n")
    service.Process(ctx, data)
}

// 2. 非 HTTP 场景手动创建
func CronTask() {
    ctx := context.Background()
    traceID := log.GetOrNewTraceID(ctx)
    ctx = context.WithValue(ctx, "trace_id", traceID)
    log.CtxInfof(ctx, "Cron started\n")
}

// 3. 调用外部服务时传递 TraceID
func CallAPI(ctx context.Context) {
    traceID := web.GetOrNewTraceID(ctx)
    req.Header.Set("X-Trace-ID", traceID)
}

// 4. GORM 使用 WithContext
db.WithContext(ctx).Where("id = ?", id).First(&user)
```

### ❌ 不推荐做法

```go
// ❌ 不传递 context
service.Process(data)

// ❌ 不使用 CtxXXXf
log.Infof("Processing\n")  // 没有 TraceID

// ❌ GORM 不使用 WithContext
db.Where("id = ?", id).First(&user)  // SQL 日志没有 TraceID

// ❌ 忘记在定时任务中设置 TraceID
func CronTask() {
    ctx := context.Background()
    log.CtxInfof(ctx, "Cron started\n")  // TraceID 为空
}
```

## 完整示例

### HTTP 请求完整链路

```go
// Handler 层
func UserLoginHandler(c *gin.Context) {
    ctx := c.Request.Context()  // TraceID 已自动注入
    
    var req LoginRequest
    c.BindJSON(&req)
    
    log.CtxInfof(ctx, "User login attempt: %s\n", req.Email)
    
    // 传递 context 到 Service 层
    result, err := userService.Login(ctx, &req)
    
    if err != nil {
        log.CtxErrorf(ctx, "Login failed: %v\n", err)
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    log.CtxInfof(ctx, "Login successful, userID: %d\n", result.UserID)
    c.JSON(200, result)
}

// Service 层
func (s *UserService) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
    log.CtxInfof(ctx, "Validating credentials\n")
    
    // 继续传递 context 到 Manager 层
    user, err := s.userManager.GetByEmail(ctx, req.Email)
    if err != nil {
        log.CtxErrorf(ctx, "User not found: %v\n", err)
        return nil, err
    }
    
    // 验证密码
    if !verifyPassword(req.Password, user.Password) {
        log.CtxWarnf(ctx, "Invalid password for user: %s\n", req.Email)
        return nil, errors.New("invalid password")
    }
    
    log.CtxInfof(ctx, "Password verified for user: %s\n", req.Email)
    
    // 生成 token
    token, _ := generateToken(user.ID)
    
    return &LoginResponse{
        UserID: user.ID,
        Token:  token,
    }, nil
}

// Manager 层
func (m *UserManager) GetByEmail(ctx context.Context, email string) (*User, error) {
    var user User
    
    // GORM 查询会自动包含 TraceID
    err := m.db.WithContext(ctx).
        Where("email = ?", email).
        First(&user).Error
    
    if err != nil {
        log.CtxErrorf(ctx, "Database query failed: %v\n", err)
        return nil, err
    }
    
    log.CtxDebugf(ctx, "User found: ID=%d, Email=%s\n", user.ID, user.Email)
    return &user, nil
}
```

### 日志输出示例

```
2026/01/16 15:30:45 user_handler.go:15: [INFO] [TraceID: 123e4567-e89b-12d3-a456-426614174000] User login attempt: user@example.com
2026/01/16 15:30:45 user_service.go:25: [INFO] [TraceID: 123e4567-e89b-12d3-a456-426614174000] Validating credentials
2026/01/16 15:30:45 gorm_logger.go:60: [INFO] [TraceID: 123e4567-e89b-12d3-a456-426614174000] [GORM] user_manager.go:45 | 1.234ms | rows:1 | SELECT * FROM `user_tab` WHERE email = 'user@example.com'
2026/01/16 15:30:45 user_manager.go:50: [DEBUG] [TraceID: 123e4567-e89b-12d3-a456-426614174000] User found: ID=1001, Email=user@example.com
2026/01/16 15:30:45 user_service.go:35: [INFO] [TraceID: 123e4567-e89b-12d3-a456-426614174000] Password verified for user: user@example.com
2026/01/16 15:30:45 user_handler.go:25: [INFO] [TraceID: 123e4567-e89b-12d3-a456-426614174000] Login successful, userID: 1001
```

## 与 RequestID 的区别

| 特性 | TraceID | RequestID |
|------|---------|-----------|
| 标准化 | ✅ 分布式追踪标准 | ⚠️ 传统命名 |
| HTTP 头 | X-Trace-ID | X-Request-ID |
| 适用场景 | HTTP + 非HTTP | 主要用于 HTTP |
| 推荐使用 | ✅ 推荐 | ⚠️ 兼容保留 |

**注意**：系统同时支持两种命名，但推荐使用 TraceID。

## 总结

- ✅ **HTTP 请求**：自动注入，直接使用 `ctx` 即可
- ✅ **定时任务**：使用 `log.GetOrNewTraceID(ctx)` 手动创建
- ✅ **异步任务**：确保传递 TraceID 到新 goroutine
- ✅ **日志输出**：使用 `log.CtxXXXf(ctx, ...)`
- ✅ **GORM 查询**：使用 `db.WithContext(ctx)`
- ✅ **外部调用**：通过 HTTP 头传递 TraceID

**核心原则**：始终传递 `context.Context`，TraceID 会自动跟随！

---

**版本**: V1.0  
**更新时间**: 2026-01-16  
**作者**: LabEquip-BMS Team
