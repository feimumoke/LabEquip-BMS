package crontask

import (
	"context"
	"encoding/json"
	"reflect"
	"runtime/debug"

	"github.com/feimumoke/labequipbms/framework/appcontext"
	"github.com/feimumoke/labequipbms/framework/bmserror"
	"github.com/feimumoke/labequipbms/framework/log"
)

type CornRunner struct {
}

func (c *CornRunner) Run() {

}

func NewCornRunner() *CornRunner {
	return &CornRunner{}
}

type TaskPayload struct {
	Param string `json:"param"`
}
type AsyncTaskHandler func(ctx context.Context) *bmserror.BMSError
type AsyncPayloadTaskHandler func(ctx context.Context, payload *TaskPayload) *bmserror.BMSError

type HandlerMethod = func(ctx context.Context, message interface{}) *bmserror.BMSError

type CronMsgTask struct {
	taskName      string
	handlerMethod []HandlerMethod
	message       interface{}
}

type SignalTaskType int64

const (
	withoutPayload SignalTaskType = 0
	withPayload    SignalTaskType = 1
)

type CronSignalTask struct {
	taskType            SignalTaskType
	taskName            string
	asyncHandler        AsyncTaskHandler
	asyncPayloadHandler AsyncPayloadTaskHandler
}

var taskMap = make(map[string]*CronMsgTask)
var signalTaskMap = make(map[string]*CronSignalTask)

// 一对多
func (c *CornRunner) RegisterMsgJob(taskName string, handler HandlerMethod, message interface{}) {
	if task, ok := taskMap[taskName]; ok {
		task.handlerMethod = append(task.handlerMethod, handler)
		//同一个taskName，message类型不同时
		if reflect.TypeOf(task.message).Elem() != reflect.TypeOf(message).Elem() {
			panic("RegisterMessageHandler message kind is different")
		}
		return
	}
	taskMap[taskName] = &CronMsgTask{
		taskName:      taskName,
		handlerMethod: []HandlerMethod{handler},
		message:       message,
	}
}

func (c *CornRunner) RegisterSignalTask(taskName string, handler AsyncTaskHandler) {
	signalTaskMap[taskName] = &CronSignalTask{
		taskType:     withoutPayload,
		taskName:     taskName,
		asyncHandler: handler,
	}
}

func (c *CornRunner) RegisterSignalPayloadTask(taskName string, handler AsyncPayloadTaskHandler) {
	signalTaskMap[taskName] = &CronSignalTask{
		taskType:            withPayload,
		taskName:            taskName,
		asyncPayloadHandler: handler,
	}
}

type SyncMessage struct {
	TaskName      string
	MessageBody   string
	SyncTime      int64
	ConsumedTimes int64 //消息被消费的次数
}
type SaturnReply struct {
	Retcode int
	Message string
}

const (
	MessageConsumeSuccess = 0
	MessageConsumeFail    = -1
)

func (b *CornRunner) HandleMsgJob(ctx context.Context, message *SyncMessage) (saturnReplay *SaturnReply) {
	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
			log.CtxErrorf(ctx, "async saturn panic error:%v, stack:%v", message, string(debug.Stack()))
			saturnReplay = &SaturnReply{Retcode: MessageConsumeFail, Message: "taskname is not exit"}
		}
	}()
	// 解析消息

	log.Infof("async saturn:%v,begin", message.TaskName)

	// 查找处理函数
	task := taskMap[message.TaskName]
	//  没有注册的task
	if task == nil {
		log.Errorf("async saturn taskname:%v is not exit", message.TaskName)
		return &SaturnReply{Retcode: MessageConsumeFail, Message: "taskname is not exit"}
	}

	for _, handlerMethod := range task.handlerMethod {
		// panic捕获是不是应该放在这
		ct := appcontext.BindContext(ctx)

		request := reflect.New(reflect.TypeOf(task.message).Elem()).Interface()
		err := json.Unmarshal([]byte(message.MessageBody), request)
		if err != nil {
			log.Errorf("async saturn json unmarshal fail:%v", task.taskName)
		}
		res := handlerMethod(ct, request)
		if res != nil {
			log.Errorf("async saturn:%v, times:%v, err:%v,", message.TaskName, message.ConsumedTimes, res.DebugError())
			return &SaturnReply{Retcode: MessageConsumeFail, Message: "task is suceess"}
		}
	}
	log.Infof("async saturn:%v,end", message.TaskName)
	return &SaturnReply{Retcode: MessageConsumeSuccess, Message: "task is suceess"}
}

func (b *CornRunner) HandleSignalJob(ctx context.Context, message *SyncMessage) (saturnReplay *SaturnReply) {
	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
			log.CtxErrorf(ctx, "signal task  panic error:%v, stack:%v", message, string(debug.Stack()))
			saturnReplay = &SaturnReply{Retcode: MessageConsumeFail, Message: "taskname is not exit"}
		}
	}()

	log.Infof("signal task :%v,begin", message.TaskName)

	// 查找处理函数
	task := signalTaskMap[message.TaskName]
	//  没有注册的task
	if task == nil {
		log.Errorf("async saturn taskname:%v is not exit", message.TaskName)
		return &SaturnReply{Retcode: MessageConsumeFail, Message: "taskname is not exit"}
	}

	if task.taskType == withoutPayload {
		res := task.asyncHandler(ctx)
		if res != nil {
			log.Errorf("signal task :%v, times:%v, err:%v,", message.TaskName, message.ConsumedTimes, res.DebugError())
			return &SaturnReply{Retcode: MessageConsumeFail, Message: "signal task is suceess"}
		}
	} else {
		payload := &TaskPayload{
			Param: message.MessageBody,
		}
		res := task.asyncPayloadHandler(ctx, payload)
		if res != nil {
			log.Errorf("signal task :%v,with payload %v times:%v, err:%v,", message.TaskName, message.MessageBody, message.ConsumedTimes, res.DebugError())
			return &SaturnReply{Retcode: MessageConsumeFail, Message: "signal task is suceess"}
		}
	}
	log.Infof("signal task :%v,end", message.TaskName)
	return &SaturnReply{Retcode: MessageConsumeSuccess, Message: "task is suceess"}
}
