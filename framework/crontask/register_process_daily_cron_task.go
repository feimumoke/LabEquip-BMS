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
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/robfig/cron"
)

const (
	APICronCatModule = "APICron.Outbound"
	daySecond        = 3600 * 24
)

// 按照从每天00：00开始偏移的秒数开始执行
type DailyCronInProcess struct {
	inner                 *cron.Cron
	dailyCronInProcessMap map[string]*DailyCronInProcessTask
	//mutex                sync.Mutex
}

type DailyCronInProcessTask struct {
	HandleName    string
	HandlerMethod CronHandlerMethod
	Message       interface{}
	ReentryLock   *int64
	Schedule      *DailySecondSchedule
}

func InitDailyCronInProcess() *DailyCronInProcess {
	return &DailyCronInProcess{
		inner:                 cron.New(),
		dailyCronInProcessMap: make(map[string]*DailyCronInProcessTask),
	}
}

func (c *DailyCronInProcess) DailyCronInProcessRun() {
	c.inner.Start()
}

func (c *DailyCronInProcess) DailyCronInProcessStop() {
	c.inner.Stop()
}

// 每天0点开始计算 taskname不可重复
func (r *DailyCronInProcess) RegisterDayOffsetByDateCronHandler(taskName string, handler CronHandlerMethod, message interface{}, second int64) {
	handlerName := getMethodName(handler)
	task := &DailyCronInProcessTask{HandleName: handlerName, HandlerMethod: handler, Message: message,
		Schedule: &DailySecondSchedule{interval: second}, ReentryLock: convert.Int64(0)}
	r.dailyCronInProcessMap[taskName] = task
	r.inner.Schedule(task.Schedule, &DailyJob{
		name: taskName,
		task: task,
	})
}

// 更新时间周期
func (r *DailyCronInProcess) ChangeMethodCronInterval(taskName string, newSecond int64) {
	if task, ok := r.dailyCronInProcessMap[taskName]; ok {
		task.Schedule.interval = newSecond
	}
}

type DailySecondSchedule struct {
	interval int64
}

// 根据当前时间到零时的时间以及时间间隔确认下一次执行时间，保证每个实例的执行时间相同
func (s *DailySecondSchedule) Next(t time.Time) time.Time {
	startTime := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	start := startTime.Unix()
	now := t.Unix()
	nextInterval := (((now - start) / s.interval) + 1) * s.interval
	var next int64
	if nextInterval > daySecond {
		next = start + daySecond
	} else {
		next = nextInterval + start
	}
	log.Infof(fmt.Sprintf("start: %v ,now: %v, next: %v", start, now, next))
	return t.Add(time.Duration(next-now) * time.Second)
}

type DailyJob struct {
	name string
	task *DailyCronInProcessTask
}

const CronTaskModule = "CronTask"

func (d *DailyJob) Run() {
	task := d.task
	handle := task.HandlerMethod
	message := task.Message
	handlerName := task.HandleName
	reentryLock := task.ReentryLock
	run := func() {
		ctx := context.Background()
		ctx = appcontext.BindContext(ctx)
		ct, endFunc := monitor.AwesomeStart1(ctx)
		var err *bmserror.BMSError
		status := 0
		if reentryLock != nil && atomic.LoadInt64(reentryLock) == 1 {
			endFunc(APICronCatModule, d.name+"."+handlerName, status, "previous cron timeout")
			return
		}
		defer func() {
			if err1 := recover(); err1 != nil {
				errMsg := string(debug.Stack())
				_ = monitor.AwesomeReportEventWithoutTrans(ct, CronTaskModule, d.name, "-1", errMsg)
				log.CtxErrorf(ct, "cron api panic error:%v, taskName:%v, taskHandle:%v, stack:%v", err1, d.name, handlerName, errMsg)
				err = bmserror.NewError(constant.ErrInternalServer, "cron api panic: %v", err1)
			}
			if reentryLock != nil {
				atomic.StoreInt64(reentryLock, 0)
			}
			trace.UnsetCtxTraceID(ct)
			endFunc(APICronCatModule, d.name+"."+handlerName, status, err.DebugError())
		}()
		hostInfo := hostutil.GetHostInfo()
		ct = trace.SetCtxTraceID(ct, hostInfo.HostName+"#"+handlerName)
		if reentryLock != nil {
			atomic.AddInt64(reentryLock, 1)
		}
		log.CtxInfof(ct, "api cron task begin")
		handlerErr := handle(ct, message)
		log.CtxInfof(ct, "api cron task end")
		if handlerErr != nil {
			log.Errorf("cron api in process err:%v, err:%v,", d.name, handlerErr.DebugError())
			status = -1
			err = handlerErr
		}
	}
	run()
}
