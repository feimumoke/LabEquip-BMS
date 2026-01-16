# TraceID 功能更新说明

## 更新概述

已将 RequestID 升级为 TraceID，符合分布式追踪标准，并添加 `GetOrNewTraceID` 函数。

## 主要变更

### 1. 命名升级

| 原名称 | 新名称 | 状态 |
|-------|--------|------|
| RequestID | TraceID | ✅ 推荐使用 |
| X-Request-ID | X-Trace-ID | ✅ 推荐使用 |
| request_id | trace_id | ✅ Context Key |

**注意**：旧名称仍然兼容，不影响现有代码。

### 2. 新增函数

#### `GetOrNewTraceID(ctx context.Context) string`

**位置**：
- `framework/web/request_id_middleware.go`
- `framework/log/logger.go`

**功能**：
- 从 context 中获取 TraceID，如果不存在则自动生成新的
- 适用于任何需要 TraceID 的场景（HTTP、定时任务、消息队列等）

**使用示例**：

```go
// 定时任务
func CronTask() {
    ctx := context.Background()
    traceID := log.GetOrNewTraceID(ctx)
    ctx = context.WithValue(ctx, "trace_id", traceID)
    
    log.CtxInfof(ctx, "Cron task started\n")
}

// 消息队列
func ConsumeMessage(msg *Message) {
    ctx := context.Background()
    traceID := log.GetOrNewTraceID(ctx)
    ctx = context.WithValue(ctx, "trace_id", traceID)
    
    log.CtxInfof(ctx, "Processing message\n")
}
```

### 3. 日志输出变更

#### 旧格式
```
[INFO] [RequestID: xxx-xxx-xxx] User login successful
```

#### 新格式
```
[INFO] [TraceID: xxx-xxx-xxx] User login successful
```

### 4. HTTP 头支持

系统同时支持两种 HTTP 头：

```bash
# 推荐使用
curl -H "X-Trace-ID: my-trace-id" http://localhost:8080/api/users

# 兼容旧名称
curl -H "X-Request-ID: my-request-id" http://localhost:8080/api/users
```

响应头：
```
X-Trace-ID: 123e4567-e89b-12d3-a456-426614174000
X-Request-ID: 123e4567-e89b-12d3-a456-426614174000
```

## 向后兼容性

✅ **完全向后兼容**，不需要修改任何现有代码！

### 兼容的函数

```go
// 这些函数仍然可用，实际返回 TraceID
web.GetRequestID(ctx)
web.GetRequestIDFromGin(c)
```

### 兼容的 HTTP 头

```
X-Request-ID  // 兼容
X-Trace-ID    // 推荐
```

### 兼容的 Context Key

```go
ctx.Value("request_id")  // 仍然可用
ctx.Value("trace_id")    // 推荐
```

## 代码迁移建议

### 可选迁移（推荐）

#### 1. 更新日志查询

```bash
# 旧方式（仍然有效）
grep "\[RequestID:" logs/bms-$(date +%Y-%m-%d).log

# 新方式（推荐）
grep "\[TraceID:" logs/bms-$(date +%Y-%m-%d).log
```

#### 2. 更新函数调用

```go
// 旧方式（仍然有效）
requestID := web.GetRequestID(ctx)

// 新方式（推荐）
traceID := web.GetTraceID(ctx)

// 或使用新函数（推荐）
traceID := web.GetOrNewTraceID(ctx)
```

#### 3. 更新 HTTP 头

```go
// 旧方式（仍然有效）
req.Header.Set("X-Request-ID", requestID)

// 新方式（推荐）
req.Header.Set("X-Trace-ID", traceID)

// 或同时设置（最佳）
traceID := web.GetOrNewTraceID(ctx)
req.Header.Set("X-Trace-ID", traceID)
req.Header.Set("X-Request-ID", traceID)  // 兼容
```

## 更新的文件

### 核心文件

1. **framework/web/request_id_middleware.go**
   - 添加 `GetOrNewTraceID` 函数
   - 添加 `GetTraceID` 和 `GetTraceIDFromGin` 函数
   - 支持 `X-Trace-ID` HTTP 头
   - 保留兼容函数

2. **framework/log/logger.go**
   - 添加 `GetOrNewTraceID` 函数
   - 添加 `generateTraceID` 函数
   - 更新日志输出格式（RequestID → TraceID）
   - 更新 `extractTraceID` 函数

### 文档文件

3. **TraceID使用指南.md**（新建）
   - 详细的使用说明
   - 多种场景示例
   - 最佳实践

4. **TraceID功能更新说明.md**（本文档）
   - 更新说明
   - 迁移指南

## 使用场景对比

### HTTP 请求（无需修改）

```go
// 之前的代码仍然有效
func Handler(c *gin.Context) {
    ctx := c.Request.Context()
    log.CtxInfof(ctx, "Processing\n")
}
```

### 定时任务（建议使用新函数）

```go
// 旧方式（需要手动生成 UUID）
func CronTask() {
    requestID := uuid.New().String()
    ctx := context.WithValue(context.Background(), "request_id", requestID)
    log.CtxInfof(ctx, "Cron started\n")
}

// 新方式（更简洁）
func CronTask() {
    ctx := context.Background()
    traceID := log.GetOrNewTraceID(ctx)
    ctx = context.WithValue(ctx, "trace_id", traceID)
    log.CtxInfof(ctx, "Cron started\n")
}
```

## TraceID 生成格式

### HTTP 请求

使用 UUID v4：
```
123e4567-e89b-12d3-a456-426614174000
```

### 非 HTTP 场景（使用 generateTraceID）

格式：`timestamp-randomhex`

示例：
```
1737028845123456789-a1b2c3d4e5f6g7h8
```

## 测试验证

### 1. HTTP 请求测试

```bash
# 测试 X-Trace-ID
curl -i -H "X-Trace-ID: test-trace-id-123" http://localhost:8080/api/enum/list

# 检查响应头
# X-Trace-ID: test-trace-id-123
# X-Request-ID: test-trace-id-123

# 查看日志
grep "test-trace-id-123" logs/bms-$(date +%Y-%m-%d).log
```

### 2. 定时任务测试

```go
// test_trace.go
package main

import (
    "context"
    "github.com/feimumoke/labequipbms/framework/log"
)

func main() {
    // 初始化日志系统
    log.InitLogger(&log.LogConfig{
        Dir:           "./logs",
        Level:         "INFO",
        EnableConsole: true,
    })
    
    // 测试 GetOrNewTraceID
    ctx := context.Background()
    traceID := log.GetOrNewTraceID(ctx)
    ctx = context.WithValue(ctx, "trace_id", traceID)
    
    log.CtxInfof(ctx, "Test started\n")
    log.CtxInfof(ctx, "TraceID: %s\n", traceID)
}
```

运行：
```bash
go run test_trace.go
tail logs/bms-$(date +%Y-%m-%d).log
```

## 常见问题

### Q1: 需要修改现有代码吗？

**A**: 不需要！所有旧代码完全兼容。但建议在新代码中使用 TraceID 命名。

### Q2: RequestID 还能用吗？

**A**: 能！`GetRequestID` 函数仍然可用，只是内部返回的是 TraceID。

### Q3: 日志查询需要改吗？

**A**: 建议更新为 `grep "\[TraceID:"` 但 `grep "\[RequestID:"` 不再有效（因为日志格式已改变）。

### Q4: HTTP 头用哪个？

**A**: 推荐使用 `X-Trace-ID`，但系统同时支持 `X-Request-ID` 以保持兼容。

### Q5: GetOrNewTraceID 和 GetTraceID 的区别？

**A**:
- `GetOrNewTraceID`: 如果不存在则创建新的（推荐用于定时任务等场景）
- `GetTraceID`: 如果不存在返回空字符串（用于检查是否已有 TraceID）

## 迁移检查清单

- [ ] 了解新的 `GetOrNewTraceID` 函数
- [ ] 更新日志查询命令（可选）
- [ ] 在新代码中使用 TraceID 命名（推荐）
- [ ] 测试 HTTP 请求的 TraceID
- [ ] 测试定时任务的 TraceID（如果有）
- [ ] 更新团队文档（如果需要）

## 相关文档

- **详细使用指南**：`TraceID使用指南.md`
- **原 RequestID 说明**：`RequestID追踪功能说明.md`
- **原 RequestID 实现**：`RequestID功能实现总结.md`
- **日志系统说明**：`日志系统说明.md`
- **快速参考**：`日志系统快速参考.md`

## 总结

✅ **新增功能**：`GetOrNewTraceID` 函数  
✅ **命名标准化**：RequestID → TraceID  
✅ **向后兼容**：所有旧代码无需修改  
✅ **更强大**：支持更多使用场景  
✅ **更标准化**：符合分布式追踪规范  

**推荐做法**：在新代码中使用 TraceID 命名和 `GetOrNewTraceID` 函数！

---

**更新时间**: 2026-01-16  
**版本**: V2.0  
**作者**: LabEquip-BMS Team
