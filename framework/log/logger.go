package log

import (
	"context"
	"fmt"
)

type Logger interface {
	Print(v ...interface{})
	CtxLogInfof(ctx context.Context, format string, v ...interface{})
	CtxLogDebugf(ctx context.Context, format string, v ...interface{})
	CtxLogErrorf(ctx context.Context, format string, v ...interface{})
	CtxLogFatalf(ctx context.Context, format string, v ...interface{})
}

func Infof(format string, v ...interface{}) {
	fmt.Printf(fmt.Sprintf(format, v))
	//if isInfoEnabled() {
	//	sp := asmutil.GetRetAddr()
	//	gLogger.doCtxLogf(nil, sp, loggerDetailChan, "INFO", format, v...)
	//}
}

func CtxInfof(ctx context.Context, format string, v ...interface{}) {
	fmt.Printf(fmt.Sprintf(format, v))
	//if isInfoEnabled() {
	//	sp := asmutil.GetRetAddr()
	//	gLogger.doCtxLogf(nil, sp, loggerDetailChan, "INFO", format, v...)
	//}
}

func CtxErrorf(ctx context.Context, format string, v ...interface{}) {
	fmt.Printf(fmt.Sprintf(format, v))
	//if isInfoEnabled() {
	//	sp := asmutil.GetRetAddr()
	//	gLogger.doCtxLogf(nil, sp, loggerDetailChan, "INFO", format, v...)
	//}
}
func Errorf(format string, v ...interface{}) {
	fmt.Printf(fmt.Sprintf(format, v))
	//if isInfoEnabled() {
	//	sp := asmutil.GetRetAddr()
	//	gLogger.doCtxLogf(nil, sp, loggerDetailChan, "INFO", format, v...)
	//}
}
