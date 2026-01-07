package asynctask

import (
	"context"
	"encoding/json"
	"fmt"

	"math/rand"
	"reflect"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/feimumoke/labequipbms/defines/constant"
	"github.com/feimumoke/labequipbms/framework/appcontext"
	"github.com/feimumoke/labequipbms/framework/bmserror"
	"github.com/feimumoke/labequipbms/framework/crontask"
	"github.com/feimumoke/labequipbms/framework/log"
	"github.com/feimumoke/labequipbms/framework/support/copier"
	"github.com/feimumoke/labequipbms/framework/support/monitor"
	"github.com/feimumoke/labequipbms/framework/support/trace"
	"github.com/feimumoke/labequipbms/framework/transaction"
)

/**

!!!!!!!!ATTENTION!!!!!!!

*/

// 发送本进程异步消息
func SendMessageInProcess(ctx context.Context, taskName string, message interface{}) *bmserror.BMSError {
	return sendMessageInProcess(ctx, taskName, message, true)
}

// 发送本进程异步消息 - 不保存
func SendMessageInProcessWithNoStore(ctx context.Context, taskName string, message interface{}) *bmserror.BMSError {
	return sendMessageInProcess(ctx, taskName, message, false)
}

// 同步调用
func SyncMessageInProcess(ctx context.Context, taskName string, message interface{}) *bmserror.BMSError {
	return syncMessageInProcess(ctx, taskName, message)
}

const (
	UndefinedInProcess int64 = 0
	NotInProcess       int64 = 1
	IsInProcess        int64 = 2
)

var inProcess int64 = UndefinedInProcess

type AsyncRunner struct {
	runner *crontask.CornRunner
}

func InitAsyncRunner(runner *crontask.CornRunner) *AsyncRunner {
	if inProcess != UndefinedInProcess {
		panic("InitAsyncRunner fail. inProcess is not UndefinedInProcess")
	}
	inProcess = NotInProcess
	r := &AsyncRunner{
		runner: runner,
	}
	return r
}

func InitAsyncRunnerInProcess(runner *crontask.CornRunner) *AsyncRunner {
	if inProcess != UndefinedInProcess {
		panic("InitAsyncRunner fail. inProcess is not UndefinedInProcess")
	}
	inProcess = IsInProcess
	return &AsyncRunner{
		runner: runner,
	}
}

// cron main
func (r *AsyncRunner) Run() {
	r.runner.Run()
}

type HandlerMethod = func(ctx context.Context, message interface{}) *bmserror.BMSError

// 注册消息处理,异步消息不能有逻辑失败
func (r *AsyncRunner) RegisterMessageHandler(taskName string, handler HandlerMethod, message interface{}) {
	if inProcess == UndefinedInProcess {
		panic("AsyncRunner has not init")
	}
	if inProcess == NotInProcess {
		r.registerMessageHandler(taskName, handler, message)
		r.registerMessageHandlerInProcess(taskName, handler, message)
	} else if inProcess == IsInProcess {
		r.registerMessageHandlerInProcess(taskName, handler, message)
	} else {
		panic("inProcess value is not in range")
	}
}

// 注册本进程内异步
func (r *AsyncRunner) RegisterMessageHandlerInProcess(taskName string, handler HandlerMethod, message interface{}) {
	r.registerMessageHandlerInProcess(taskName, handler, message)
}

// 注册定时任务
func (r *AsyncRunner) RegisterCrontabHandler(taskName string, handler crontask.AsyncTaskHandler) {
	r.registerCrontabHandler(taskName, handler)
}

// 注册定时任务
func (r *AsyncRunner) RegisterPayloadCrontabHandler(taskName string, payloadHandler crontask.AsyncPayloadTaskHandler) {
	r.registerPayloadCrontabHandler(taskName, payloadHandler)
}

// -------***------以下都是实现部分,可以不用关注-------***-------

// ---**---本进程内异步任务发送---**---

func rand8String() string {
	n := 4
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	s := fmt.Sprintf("%X", b)
	return s
}

type TaskRunner struct {
	TaskName string
	Count    int64
	Ch       chan struct{}
}

var taskRunnerMap = make(map[string]*TaskRunner)
var taskRunnerMutex sync.Mutex

func sendMessageInProcess(ctx context.Context, taskName string, message interface{}, store bool) *bmserror.BMSError {
	task, ok := ProcessTaskMap[taskName]
	if !ok {
		// 系统没有注册对应的消息处理方法
		return nil
	}

	if reflect.TypeOf(message).Kind() != reflect.Ptr {
		return bmserror.NewError(constant.ErrParam, "message kind is not ptr")
	}

	var extraData *transaction.ExtraData
	extraDataT := ctx.Value(transaction.ExtraDataKey)
	//在事务中发送消息
	if extraDataT != nil {
		extraData = extraDataT.(*transaction.ExtraData)
	}

	reqID := trace.GetOrNewTraceID(ctx)
	messageUUID := splitReqID(reqID) + "_" + rand8String()
	// 消息落地
	if store {
		buf, _ := json.Marshal(message)
		processMessage := &MProcessMessage{
			MessageUUID: messageUUID,
			TaskName:    taskName,
			Message:     string(buf),
			CTime:       time.Now().Unix(),
		}
		handlerNames := make([]string, 0, len(task.HandlerMethod))
		for k := range task.HandlerMethod {
			handlerNames = append(handlerNames, k)
		}
		errC := CreateProcessMessage(ctx, processMessage, handlerNames)
		if errC != nil {
			return errC.Mark()
		}
	}

	// 异步处理
	go func() {
		if extraData != nil { //在事务中
			atomic.AddInt64(&extraData.NeedSendCount, 1)
			success := true
			if extraData.InTransactionCount != 0 {
				select {
				case <-extraData.CompleteCh:
					atomic.AddInt64(&extraData.NeedSendCount, -1)
					if extraData.NeedSendCount > 0 {
						extraData.CompleteCh <- struct{}{}
					}
					success = true
				case <-extraData.FailCh:
					atomic.AddInt64(&extraData.NeedSendCount, -1)
					if extraData.NeedSendCount > 0 {
						extraData.FailCh <- struct{}{}
					}
					success = false
				}
			} else {
				atomic.AddInt64(&extraData.NeedSendCount, -1)
			}
			if !success {
				return
			}
		}

		var consumerCount = int64(len(task.HandlerMethod))

		for name, tHandler := range task.HandlerMethod {
			handlerName := name
			handler := tHandler
			go func() {

				ct := context.Background()
				ct = trace.SetCtxTraceID(ct, newReqIDWithHandleName(reqID, handlerName))

				taskRunnerMutex.Lock()
				taskRunner, ok := taskRunnerMap[taskName]
				if !ok {
					taskRunner = &TaskRunner{
						TaskName: taskName,
						Count:    0,
						Ch:       make(chan struct{}, 30),
					}
					taskRunnerMap[taskName] = taskRunner
				}
				taskRunner.Count++
				taskRunnerMutex.Unlock()

				_, endFunc2 := monitor.AwesomeStart1(ct)
				taskRunner.Ch <- struct{}{}
				endFunc2("AsyncInProcessWait", taskName, 0, "")

				defer func() {
					taskRunnerMutex.Lock()
					t := taskRunnerMap[taskName]
					t.Count--
					taskRunnerMutex.Unlock()
					<-t.Ch
				}()

				var err *bmserror.BMSError

				defer func() {
					if err1 := recover(); err1 != nil {
						errMsg := string(debug.Stack())
						_ = monitor.AwesomeReportEventWithoutTrans(ct, CronTaskModule, taskName, "-1", errMsg)
						log.CtxErrorf(ct, "async saturn panic error:%v, taskName:%v, taskHandle:%v, stack:%v", err1, taskName, handlerName, errMsg)
						err = bmserror.NewError(constant.ErrInternalServer, "async message panic: %v", err1)
						if store {
							_ = SetMessageConsumerHandlerStatus(ct, messageUUID, handlerName, ConsumerHandlerFail)
						}
					}

					//unset log id
					trace.UnsetCtxTraceID(ct)

				}()

				// 拷贝待优化
				msg1 := reflect.New(reflect.TypeOf(task.Message).Elem()).Interface()
				_ = copier.Copy(message, msg1)

				//初始话ctx
				//业务ctx根据仓库初始化
				ct = appcontext.BindContext(ct)

				err = handler(ct, msg1)
				//获取对应仓库的连接，用于删除落库的消息
				if err != nil {
					log.Errorf("async fail, taskName:%v, handlerName:%v, err:%v,", taskName, handlerName, err.DebugError())
					if store {
						_ = SetMessageConsumerHandlerStatus(ct, messageUUID, handlerName, ConsumerHandlerFail)
					}
				} else {
					if store {
						_ = DeleteMessageConsumer(ct, messageUUID, handlerName)
						atomic.AddInt64(&consumerCount, -1)
						if consumerCount == 0 {
							_ = DeleteProcessMessage(ct, messageUUID, taskName)
						}
					}
				}
			}()
		}
	}()
	return nil
}

func newReqIDWithHandleName(reqID, HandlerName string) string {
	return reqID + "#" + HandlerName
}

func splitReqID(reqID string) string {
	s := strings.Split(reqID, "#")
	return s[0]
}

// ---**---本进程内的同步任务语法糖---**---
// 同步消息的ctx与执行事务一致，也不需要落库，因为是同时执行或者回滚
func syncMessageInProcess(ctx context.Context, taskName string, message interface{}) *bmserror.BMSError {

	task, ok := ProcessTaskMap[taskName]
	if !ok {
		// 系统没有注册对应的消息处理方法
		return nil
	}

	if reflect.TypeOf(message).Kind() != reflect.Ptr {
		return bmserror.NewError(constant.ErrParam, "message kind is not ptr")
	}

	for tHandlerName, tHandler := range task.HandlerMethod {
		handlerName := tHandlerName
		handler := tHandler
		var err *bmserror.BMSError
		func() {
			ct, endFunc := monitor.AwesomeStart1(ctx)
			status := 0

			defer func() {
				if panicErr := recover(); panicErr != nil {
					debug.PrintStack()

					log.CtxErrorf(ct, "sync message in process panic error:%v, taskName:%v, taskHandle:%v, stack:%v", panicErr, taskName, handlerName, string(debug.Stack()))
					_ = monitor.AwesomeReportEventWithoutTrans(ct, CronTaskModule, taskName, "-1", string(debug.Stack()))

					err = bmserror.NewError(constant.ErrInternalServer, "sync message in process panic error:%v", panicErr)
				}

				endFunc("SyncInProcess", task.TaskName+"."+handlerName, status, err.DebugError())
			}()

			msg1 := reflect.New(reflect.TypeOf(task.Message).Elem()).Interface()

			// 拷贝待优化
			copyErr := copier.Copy(message, msg1)
			if copyErr != nil {
				log.CtxInfof(ctx, "sync copy err :%v", copyErr)
			}

			handlerErr := handler(ct, msg1)
			if handlerErr != nil {
				log.Errorf("sync message in process err:%v, err:%v,", taskName, handlerErr.DebugError())
				status = -1
				err = handlerErr
			}
		}()

		if err != nil {
			return err.Mark()
		}
	}
	return nil
}

// ---**---本进程内异步任务注册---**---

type ProcessTask struct {
	TaskName      string
	HandlerMethod map[string]HandlerMethod
	Message       interface{}
}

var ProcessTaskMap = make(map[string]*ProcessTask)

func getMethodName(handler HandlerMethod) string {
	handlerNames := strings.Split(runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name(), "/")
	handlerName := handlerNames[len(handlerNames)-1]
	handlerName = strings.Replace(handlerName, "-fm", "", -1)
	handlerName = strings.Replace(handlerName, "(", "", -1)
	handlerName = strings.Replace(handlerName, ")", "", -1)
	handlerName = strings.Replace(handlerName, "*", "", -1)
	handlerName = strings.Replace(handlerName, ".", "_", -1)
	return handlerName
}

func (r *AsyncRunner) registerMessageHandlerInProcess(taskName string, handler HandlerMethod, message interface{}) {
	//不允许非指针注册
	t := reflect.TypeOf(message)
	if t.Kind() != reflect.Ptr {
		panic("RegisterMessageHandler message is not ptr")
	}

	handlerName := getMethodName(handler)
	wrapperAsyncTask := r.wrapperAsyncTask(taskName, handler)

	//同一个taskName有多个处理方法
	if task, ok := ProcessTaskMap[taskName]; ok {
		//同一个taskName，message类型不同时
		if reflect.TypeOf(task.Message).Elem() != reflect.TypeOf(message).Elem() {
			panic("RegisterMessageHandler message kind is different")
		}
		if _, ok := task.HandlerMethod[handlerName]; ok {
			panic(fmt.Sprintf("RegisterMessageHandler %s had registered", handlerName))
		}
		task.HandlerMethod[handlerName] = wrapperAsyncTask
		return
	}

	ProcessTaskMap[taskName] = &ProcessTask{
		TaskName:      taskName,
		HandlerMethod: map[string]HandlerMethod{handlerName: wrapperAsyncTask},
		Message:       message,
	}
}

// ---**---发送异步任务相关---**--

// ---**---异步任务实现相关---**---

// 注册消息处理,异步消息不能有逻辑失败
func (r *AsyncRunner) registerMessageHandler(taskName string, handler HandlerMethod, message interface{}) {
	//不允许非指针注册
	t := reflect.TypeOf(message)
	if t.Kind() != reflect.Ptr {
		panic("RegisterMessageHandler message is not ptr")
	}
	wrapperHandler := r.wrapperAsyncTask(taskName, handler)
	r.runner.RegisterMsgJob(taskName, wrapperHandler, message)
}

func setTraceID(ctx context.Context) context.Context {
	traceID := trace.GetOrNewTraceID(ctx)
	ctx = trace.SetCtxTraceID(ctx, traceID)
	return ctx
}

func (r *AsyncRunner) wrapperAsyncTask(taskName string, task func(ctx context.Context, message interface{}) *bmserror.BMSError) func(ctx context.Context, message interface{}) *bmserror.BMSError {
	return func(ctx context.Context, message interface{}) (err *bmserror.BMSError) {
		//set log id
		hasTraceID := trace.CtxHasTraceID(ctx)
		if !hasTraceID {
			ctx = setTraceID(ctx)
		}

		timer := time.Now().Unix()
		defer func() {
			log.CtxInfof(ctx, "asynctask %v consumed:[%v]", taskName, timer-time.Now().Unix())
			if err != nil {
				log.CtxErrorf(ctx, "AsyncTask return error: %v", err.Error())
			}
			if p := recover(); p != nil {
				//打印调用栈信息
				errMsg := string(debug.Stack())
				err = bmserror.NewError(constant.ErrInternalServer, "panic:%v", p)
				log.CtxErrorf(ctx, "AsyncTask runner inner panic: %v ,stack: %v", p, errMsg)
			}

			if !hasTraceID {
				trace.UnsetCtxTraceID(nil)
			}
		}()

		return task(ctx, message)
	}
}

// ---**---注册定时任务实现---**---

var crontabMap = make(map[string]int)

func (r *AsyncRunner) registerCrontabHandler(taskName string, handler func(ctx context.Context) *bmserror.BMSError) {
	if _, ok := crontabMap[taskName]; ok {
		panic("RegisterCrontabHandler task name has same,taskName=" + taskName)
	}
	wrapperHandle := r.wrapperCrontabTask(taskName, handler)
	r.runner.RegisterSignalTask(taskName, wrapperHandle)
}

func (r *AsyncRunner) wrapperCrontabTask(taskName string, handler func(ctx context.Context) *bmserror.BMSError) func(ctx context.Context) *bmserror.BMSError {
	return func(ctx context.Context) (err *bmserror.BMSError) {
		//set log id
		ctx = setTraceID(ctx)
		ctx, endFunc := monitor.AwesomeStart1(ctx)
		start := time.Now().Unix()
		defer func() {
			log.CtxInfof(ctx, "cronTask #%v# consumed:[%v]", taskName, time.Now().Unix()-start)

			if err != nil {
				log.CtxErrorf(ctx, "CrontabTask return error: %v", err.Error())
				endFunc("CrontabTask", taskName, -1, err.DebugError())

			} else {
				endFunc("CrontabTask", taskName, 0, "success")
			}
			if p := recover(); p != nil {
				//打印调用栈信息
				errMsg := string(debug.Stack())
				log.CtxErrorf(ctx, "CrontabTask inner panic: %v ,stack: %v", p, errMsg)
				err = bmserror.NewError(constant.ErrInternalServer, "panic:%v", p)
			}
			trace.UnsetCtxTraceID(ctx)
		}()
		log.CtxInfof(ctx, "cronTask #%v# start:%v", taskName, time.Now().Unix())

		return handler(ctx)
	}
}

func (r *AsyncRunner) registerPayloadCrontabHandler(taskName string, payloadHandler crontask.AsyncPayloadTaskHandler) {
	if _, ok := crontabMap[taskName]; ok {
		panic("RegisterCrontabHandler task name has same,taskName=" + taskName)
	}
	wrapperHandle := r.wrapperPayloadCrontabTask(taskName, payloadHandler)

	r.runner.RegisterSignalPayloadTask(taskName, wrapperHandle)
}

func (r *AsyncRunner) wrapperPayloadCrontabTask(taskName string, payloadHandler crontask.AsyncPayloadTaskHandler) crontask.AsyncPayloadTaskHandler {
	return func(ctx context.Context, payload *crontask.TaskPayload) (err *bmserror.BMSError) {
		ctx = setTraceID(ctx)
		ctx, endFunc := monitor.AwesomeStart1(ctx)
		defer func() {
			if err != nil {
				log.CtxErrorf(ctx, "PayloadCrontabTask return error: %v", err.DebugError())
				endFunc("CrontabTask", taskName, -1, err.DebugError())
			} else {
				endFunc("CrontabTask", taskName, 0, "success")
			}

			if p := recover(); p != nil {
				//打印调用栈信息
				errMsg := string(debug.Stack())
				log.CtxErrorf(ctx, "PayloadCrontabTask inner panic: %v ,stack: %v", p, errMsg)
				err = bmserror.NewError(constant.ErrInternalServer, "panic:%v", p)
			}
			trace.UnsetCtxTraceID(ctx)
		}()

		return payloadHandler(ctx, payload)
	}
}
