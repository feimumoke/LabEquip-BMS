package crontask

import (
	"context"
	"fmt"
	"github.com/feimumoke/wechating/framework/appcontext"
	"github.com/feimumoke/wechating/framework/constant"
	"github.com/feimumoke/wechating/framework/log"
	"github.com/feimumoke/wechating/framework/support/convert"
	"github.com/feimumoke/wechating/framework/support/hostutil"
	"github.com/feimumoke/wechating/framework/support/monitor"
	"github.com/feimumoke/wechating/framework/support/trace"
	"github.com/feimumoke/wechating/framework/wcerror"
	"github.com/robfig/cron"
	"reflect"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
)

var globalWaveCronInProcessTaskMap = make(map[string]*CronInProcessTask)
var taskMapRWLock = new(sync.RWMutex)

type ExpressionCronInProcess struct {
	inner                *cron.Cron
	waveCronInProcessMap map[string]*CronInProcessTask
	//mutex                sync.Mutex
}

type CronInProcessTask struct {
	HandleName    string
	HandlerMethod CronHandlerMethod
	Message       interface{}
	ReentryLock   *int64
	Cron          string
}

type CronHandlerMethod = func(ctx context.Context, message interface{}) *bmserror.BMSError

func InitWaveCronInProcess() *ExpressionCronInProcess {
	return &ExpressionCronInProcess{
		inner:                cron.New(),
		waveCronInProcessMap: make(map[string]*CronInProcessTask),
	}
}

func (c *ExpressionCronInProcess) WaveCronInProcessRun() {
	for taskName, v := range c.waveCronInProcessMap {
		handle := v.HandlerMethod
		message := v.Message
		handlerName := v.HandleName
		reentryLock := v.ReentryLock

		a := func() {
			ctx := context.Background()
			ctx = appcontext.BindContext(ctx)
			ct, endFunc := monitor.AwesomeStart1(ctx)
			var err *bmserror.BMSError
			status := 0
			if reentryLock == nil {
				return
			}
			if !atomic.CompareAndSwapInt64(reentryLock, 0, 1) {
				status = -1 //cron reentry
				log.CtxErrorf(ct, "cron wave reentry, taskName:%v, taskHandle:%v", taskName, handlerName)
				endFunc(CronTaskModule, taskName+"."+handlerName, status, "previous cron timeout")
				return
			}
			defer func() {
				if err1 := recover(); err1 != nil {
					errMsg := string(debug.Stack())
					_ = monitor.AwesomeReportEventWithoutTrans(ct, CronTaskModule, taskName, "-1", errMsg)
					log.CtxErrorf(ct, "cron wave panic error:%v, taskName:%v, taskHandle:%v, stack:%v", err1, taskName, handlerName, errMsg)
					err = bmserror.NewError(constant.ErrInternalServer, "cron wave panic: %v", err1)
				}
				if reentryLock != nil {
					atomic.StoreInt64(reentryLock, 0)
				}
				trace.UnsetCtxTraceID(ct)
				endFunc(CronTaskModule, taskName+"."+handlerName, status, err.DebugError())
			}()
			hostInfo := hostutil.GetHostInfo()
			ct = trace.SetCtxTraceID(ct, hostInfo.HostName+"#"+handlerName)
			log.CtxInfof(ct, "wave cron task begin")
			handlerErr := handle(ct, message)
			log.CtxInfof(ct, "wave cron task end")
			if handlerErr != nil {
				log.Errorf("cron wave in process err:%v, err:%v,", taskName, handlerErr.DebugError())
				status = -1
				err = handlerErr
			}
		}
		err := c.inner.AddFunc(v.Cron, a)
		if err != nil {
			log.Errorf("cron wave in process err:%v, err:%v,", taskName, err.Error())
		}
	}
	c.inner.Start()
}

func (c *ExpressionCronInProcess) WaveCronInProcessStop() {
	c.inner.Stop()
	taskMapRWLock.RLock()
	defer func() {
		taskMapRWLock.RUnlock()
	}()
	for k, _ := range c.waveCronInProcessMap {
		delete(globalWaveCronInProcessTaskMap, k)
	}
	c.waveCronInProcessMap = make(map[string]*CronInProcessTask)
}

// 注册消息处理,异步消息不能有逻辑失败
func (r *ExpressionCronInProcess) RegisterWaveInProcessCornHandlerTimeOutNotReentry(taskName string, handler CronHandlerMethod, message interface{}, cron string) {
	handlerName := getMethodName(handler)
	taskMapRWLock.Lock()
	defer func() {
		taskMapRWLock.Unlock()
	}()
	if _, ok := globalWaveCronInProcessTaskMap[taskName]; ok {
		panic("RegisterCrontabHandler task name has same,taskName=" + taskName)
	}
	globalWaveCronInProcessTaskMap[taskName] = &CronInProcessTask{HandleName: handlerName, HandlerMethod: handler, Message: message,
		Cron: cron, ReentryLock: convert.Int64(0)}
	r.waveCronInProcessMap[taskName] = globalWaveCronInProcessTaskMap[taskName]
}

// 获取handlername
func GetMethodNameByTaskName(taskName string) string {
	taskMapRWLock.RLock()
	defer func() {
		taskMapRWLock.RUnlock()
	}()
	if l, ok := globalWaveCronInProcessTaskMap[taskName]; ok {
		return l.HandleName
	}
	return ""
}

func getMethodName(handler CronHandlerMethod) string {
	fmt.Println(runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name())
	handlerNames := strings.Split(runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name(), "/")
	handlerName := handlerNames[len(handlerNames)-1]
	handlerName = strings.Replace(handlerName, "-fm", "", -1)
	handlerName = strings.Replace(handlerName, "(", "", -1)
	handlerName = strings.Replace(handlerName, ")", "", -1)
	handlerName = strings.Replace(handlerName, "*", "", -1)
	handlerName = strings.Replace(handlerName, ".", "_", -1)
	return handlerName
}
