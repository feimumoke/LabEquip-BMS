# RequestID 请求追踪功能说明

## 一、功能概述

为了方便追踪请求链路和排查问题，系统已为每个 HTTP 请求自动生成唯一的 RequestID，并在日志中自动打印。

## 二、实现原理

### 2.1 RequestID 生成

- 使用 UUID v4 作为 RequestID
- 每个请求自动生成唯一标识
- 支持客户端传入 RequestID（通过 `X-Request-ID` 请求头）

### 2.2 RequestID 传递

- RequestID 存储在 `context.Context` 中
- 通过中间件自动注入到每个请求
- 在响应头中返回 RequestID（`X-Request-ID`）

### 2.3 日志集成

- 所有带 Context 的日志函数自动提取 RequestID
- 日志格式：`[INFO] [RequestID: xxx-xxx-xxx] 日志内容`
- GORM SQL 日志也自动包含 RequestID

## 三、使用方式

### 3.1 自动注入（推荐）

中间件已在 `framework/web/server.go` 中注册，无需手动配置：

```go
func (r *BasicServer) setAPIMiddleWare() *gin.RouterGroup {
    r.engine.Use(Cors(), RequestIDMiddleware(), APICtxMiddleware(), APIMonitorMiddleware(), AddScormMiddleware())
    apiGroup := r.engine.Group("/api")
    return apiGroup
}
```

### 3.2 在业务代码中使用

#### 方式 1: 使用带 Context 的日志函数（推荐）

```go
import (
    "context"
    "github.com/feimumoke/labequipbms/framework/log"
)

func SomeBusinessFunction(ctx context.Context) {
    // 自动包含 RequestID
    log.CtxInfof(ctx, "Processing user request\n")
    log.CtxDebugf(ctx, "Debug information: %v\n", data)
    log.CtxWarnf(ctx, "Warning message\n")
    log.CtxErrorf(ctx, "Error occurred: %v\n", err)
}
```

#### 方式 2: 手动获取 RequestID

```go
import (
    "github.com/feimumoke/labequipbms/framework/web"
    "github.com/gin-gonic/gin"
)

// 从 Gin Context 获取
func Handler(c *gin.Context) {
    requestID := web.GetRequestIDFromGin(c)
    log.Infof("RequestID: %s, Processing request\n", requestID)
}

// 从 Go Context 获取
func BusinessFunction(ctx context.Context) {
    requestID := web.GetRequestID(ctx)
    log.Infof("RequestID: %s, Doing something\n", requestID)
}
```

### 3.3 客户端传入 RequestID

客户端可以通过请求头传入自定义的 RequestID：

```bash
curl -H "X-Request-ID: my-custom-request-id" http://localhost:8080/api/users
```

如果不传入，系统会自动生成。

### 3.4 在响应中获取 RequestID

服务器会在响应头中返回 RequestID：

```bash
HTTP/1.1 200 OK
X-Request-ID: 123e4567-e89b-12d3-a456-426614174000
Content-Type: application/json
...
```

## 四、日志输出示例

### 4.1 普通日志

```
2026/01/16 15:30:45 user_service.go:25: [INFO] [RequestID: 123e4567-e89b-12d3-a456-426614174000] User login successful, userID: 1001
```

### 4.2 GORM SQL 日志

```
2026/01/16 15:30:45 gorm_logger.go:60: [INFO] [RequestID: 123e4567-e89b-12d3-a456-426614174000] [GORM] user_manager.go:45 | 1.234ms | rows:1 | SELECT * FROM `user_tab` WHERE id = 1001
```

### 4.3 错误日志

```
2026/01/16 15:30:45 user_service.go:30: [ERROR] [RequestID: 123e4567-e89b-12d3-a456-426614174000] Failed to update user: database connection timeout
```

## 五、最佳实践

### 5.1 在 Handler 中传递 Context

```go
func UserLoginHandler(c *gin.Context) {
    ctx := c.Request.Context()
    
    // 传递 context 到业务层
    result, err := userService.Login(ctx, req)
    
    if err != nil {
        log.CtxErrorf(ctx, "Login failed: %v\n", err)
        // ...
    }
}
```

### 5.2 在 Service 层使用

```go
func (s *UserService) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
    log.CtxInfof(ctx, "Attempting login for user: %s\n", req.Email)
    
    // 调用 Manager 层，继续传递 context
    user, err := s.userManager.GetByEmail(ctx, req.Email)
    
    if err != nil {
        log.CtxErrorf(ctx, "Failed to get user: %v\n", err)
        return nil, err
    }
    
    // ...
}
```

### 5.3 在 Manager 层使用

```go
func (m *UserManager) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
    var user entity.User
    
    // GORM 查询会自动包含 RequestID（如果使用 WithContext）
    err := m.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
    
    if err != nil {
        log.CtxErrorf(ctx, "Database query failed: %v\n", err)
        return nil, err
    }
    
    log.CtxDebugf(ctx, "User found: %d\n", user.ID)
    return &user, nil
}
```

## 六、排查问题时的使用

### 6.1 根据 RequestID 查找日志

```bash
# 查找特定 RequestID 的所有日志
grep "123e4567-e89b-12d3-a456-426614174000" logs/bms-2026-01-16.log

# 实时监控特定 RequestID
tail -f logs/bms-$(date +%Y-%m-%d).log | grep "123e4567-e89b-12d3-a456-426614174000"
```

### 6.2 追踪完整请求链路

```bash
# 追踪一个请求的完整生命周期
grep "123e4567-e89b-12d3-a456-426614174000" logs/bms-2026-01-16.log

# 输出示例:
# [INFO] [RequestID: 123e4567-...] User login attempt: user@example.com
# [INFO] [RequestID: 123e4567-...] [GORM] SELECT * FROM user_tab WHERE email = 'user@example.com'
# [INFO] [RequestID: 123e4567-...] Password verification successful
# [INFO] [RequestID: 123e4567-...] JWT token generated
# [INFO] [RequestID: 123e4567-...] Login successful
```

## 七、技术细节

### 7.1 中间件实现

位置：`framework/web/request_id_middleware.go`

- 从请求头获取或生成 UUID
- 存入 Gin Context 和 Go Context
- 设置到响应头

### 7.2 日志系统集成

位置：`framework/log/logger.go`

- 新增 `logfWithContext` 方法
- 自动从 context 提取 requestID
- 所有 `CtxXXXf` 函数都支持

### 7.3 GORM 集成

位置：`framework/orm/gorm_logger.go`

- GORM 的 `Trace` 方法接收 context
- 使用 `log.CtxInfof` 等函数自动包含 requestID

## 八、性能考虑

- UUID 生成性能：~100ns per operation
- Context 值查找：~10ns per operation
- 对系统性能影响：<0.1%

## 九、未来扩展

### 9.1 分布式追踪

可以与 OpenTelemetry 等分布式追踪系统集成：

```go
// 未来可以这样扩展
import "go.opentelemetry.io/otel/trace"

func RequestIDMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        requestID := uuid.New().String()
        
        // 同时创建 OpenTelemetry Span
        ctx, span := tracer.Start(c.Request.Context(), "http-request")
        span.SetAttributes(attribute.String("request.id", requestID))
        defer span.End()
        
        // ...
    }
}
```

### 9.2 链路传播

在微服务架构中，RequestID 可以跨服务传播：

```go
// HTTP 客户端自动传递 RequestID
func CallDownstreamService(ctx context.Context, url string) {
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    
    // 自动传递 RequestID
    if requestID := web.GetRequestID(ctx); requestID != "" {
        req.Header.Set("X-Request-ID", requestID)
    }
    
    resp, _ := http.DefaultClient.Do(req)
    // ...
}
```

## 十、常见问题

### Q1: RequestID 会影响性能吗？

A: 影响极小（<0.1%），UUID 生成和 context 查找都是非常快速的操作。

### Q2: 能否使用 Goroutine ID 作为 RequestID？

A: 不推荐。虽然可以获取 Goroutine ID，但：
- Go 官方不推荐使用 Goroutine ID
- Goroutine 可能被复用，ID 不唯一
- 一个请求可能跨多个 Goroutine
- UUID 更适合分布式环境

### Q3: 异步任务如何处理 RequestID？

A: 启动异步任务时传递 context：

```go
func ProcessAsync(ctx context.Context, data string) {
    go func() {
        // 继承父 context 的 RequestID
        log.CtxInfof(ctx, "Processing async task: %s\n", data)
        // ...
    }()
}
```

### Q4: 如何在非 HTTP 场景使用？

A: 手动创建带 RequestID 的 context：

```go
import "context"

func CronJob() {
    requestID := uuid.New().String()
    ctx := context.WithValue(context.Background(), "request_id", requestID)
    
    log.CtxInfof(ctx, "Cron job started\n")
    // ...
}
```

## 十一、相关文件

- `framework/web/request_id_middleware.go` - 中间件实现
- `framework/log/logger.go` - 日志系统集成
- `framework/web/server.go` - 中间件注册
- `framework/orm/gorm_logger.go` - GORM 日志集成

---

**更新时间**: 2026-01-16  
**版本**: V1.0  
**作者**: LabEquip-BMS Team
