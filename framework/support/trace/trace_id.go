package trace

import (
	"context"
	"github.com/google/uuid"
	"os"
	"strings"
)

const gLogIdKey = "logid"

type UnsetTraceFunc func(ctx context.Context)

func GetOrNewTraceID(ctx context.Context) string {
	//ctx
	if traceID, exist := getTraceIDFromCtx(ctx); exist {
		return traceID
	}

	//local map
	if traceID, exist := getTraceIDFromLocalMap(); exist {
		return traceID
	}

	return genNewTraceID()
}

func unsetTraceIDFunc(ctx context.Context) {
	unsetTraceID()
}

// important !!!  一定要调用
func UnsetCtxTraceID(ctx context.Context) {
	unsetTraceID()
}

func Init(ctx context.Context) (context.Context, UnsetTraceFunc) {
	traceID := genNewTraceID()
	setLogTraceID(traceID)

	ctx = setCtxTraceID(ctx, traceID)
	return ctx, unsetTraceIDFunc
}

func SetCtxTraceID(ctx context.Context, traceID string) context.Context {
	setLogTraceID(traceID)

	ctx = setCtxTraceID(ctx, traceID)
	return ctx
}

func InitWithTraceID(ctx context.Context, traceID string) (context.Context, UnsetTraceFunc) {
	setLogTraceID(traceID)

	ctx = setCtxTraceID(ctx, traceID)
	return ctx, unsetTraceIDFunc
}

func getTraceIDFromCtx(ctx context.Context) (string, bool) {
	var reqID string
	if ctx == nil {
		return "", false
	}
	if logID, exist := ctx.Value(gLogIdKey).(string); exist {
		reqID = logID
	}

	return reqID, len(reqID) > 0
}

func genNewTraceID() string {
	traceID := uuid.New().String()
	traceID = strings.ReplaceAll(traceID, "-", "")

	prefix := getTracePrefix()

	return prefix + getCID() + traceID
}

func getTracePrefix() string {
	prefix := ""
	module := getEnv("MODULE_NAME")
	if strings.Contains(module, "api") {
		prefix = "API"
	}
	if strings.Contains(module, "cron") {
		prefix = "CRON"
	}

	return prefix
}

func CtxHasTraceID(ctx context.Context) bool {
	_, exist := getTraceIDFromCtx(ctx)
	return exist
}

func setCtxTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, gLogIdKey, traceID)
}

func getCID() string {
	return getEnv("CID")
}

func getEnv(env string) string {
	return strings.ToUpper(os.Getenv(env))
}
