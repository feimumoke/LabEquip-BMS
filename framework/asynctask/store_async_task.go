package asynctask

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/feimumoke/labequipbms/defines/constant"
	"github.com/feimumoke/labequipbms/framework/bmserror"
	"github.com/feimumoke/labequipbms/framework/crontask"
	"github.com/feimumoke/labequipbms/framework/datasource"
	"github.com/feimumoke/labequipbms/framework/log"
	"github.com/feimumoke/labequipbms/framework/orm"
	"github.com/feimumoke/labequipbms/framework/support/monitor"
	"github.com/feimumoke/labequipbms/framework/web"

	"reflect"
	"runtime/debug"
	"time"
)

type MProcessMessage struct {
	ID          int64  `gorm:"column:id;primary_key" json:"id"`
	MessageUUID string `gorm:"column:message_uuid" json:"message_uuid"`
	TaskName    string `gorm:"column:task_name" json:"task_name"`
	Message     string `gorm:"column:message" json:"message"` //json序列化消息
	CTime       int64  `gorm:"column:ctime" json:"ctime"`
}

const ProcessMessageTable = "process_message_v2_tab"

type MMessageConsumer struct {
	ID            int64  `gorm:"column:id;primary_key" json:"id"`
	MessageUUID   string `gorm:"column:message_uuid" json:"message_uuid"`
	HandlerName   string `gorm:"column:handler_name" json:"handler_name"`
	HandlerStatus int32  `gorm:"column:handler_status" json:"handler_status"`
	HandlerTimes  int32  `gorm:"column:handler_times" json:"handler_times"`
	CTime         int64  `gorm:"column:ctime" json:"ctime"`
	MTime         int64  `gorm:"column:mtime" json:"mtime"`
}

type MRetryMessage struct {
	*MMessageConsumer
	TaskName string `gorm:"column:task_name" json:"task_name"`
	Message  string `gorm:"column:message" json:"message"`
}

const MessageConsumerTable = "process_message_consumer_v2_tab"

const (
	ConsumerHandlerFail                   = -1
	ConsumerHandlerFailTaskNotFound       = -2
	ConsumerHandlerFailTaskMsgDecodeError = -3
	ConsumerHandlerException              = -4
	ConsumerHandlerPending                = 0
	ConsumerHandlerSuccess                = 1
	ConsumerHandlerOngoing                = 2
)

const (
	MessageRetryTable = `
process_message_v2_tab as a
inner join 
process_message_consumer_v2_tab as b
on a.message_uuid = b.message_uuid
`
	MessageRetryColumn = `
a.task_name, a.message, 
b.*
`
)

var messageDs = datasource.DefaultBMSSource

func GetMessageOrm(ctx context.Context) orm.GORM {
	return messageDs.GetDataSource(ctx, nil)

}

var gFromId int64

func CreateProcessMessage(ctx context.Context, m *MProcessMessage, handlerNames []string) *bmserror.BMSError {
	cl := make([]*MMessageConsumer, 0, len(handlerNames))
	for _, s := range handlerNames {
		m := &MMessageConsumer{
			ID:            0,
			MessageUUID:   m.MessageUUID,
			HandlerName:   s,
			HandlerStatus: 0,
			HandlerTimes:  0,
			CTime:         m.CTime,
			MTime:         m.CTime,
		}
		cl = append(cl, m)
	}

	err := GetMessageOrm(ctx).Table(ProcessMessageTable).Create(m).GetError()
	if err != nil {
		return bmserror.NewError(constant.ErrInternalServer, err.Error())
	}
	err1 := GetMessageOrm(ctx).Table(MessageConsumerTable).Create(cl).GetError()
	if err1 != nil {
		return err1.Mark()
	}
	return nil
}

func DeleteProcessMessage(ctx context.Context, messageUUID, taskName string) *bmserror.BMSError {
	err := GetMessageOrm(ctx).Table(ProcessMessageTable).
		Where("message_uuid = ? and task_name = ?", messageUUID, taskName).Delete(&MProcessMessage{}).GetError()
	if err != nil {
		return bmserror.NewError(constant.ErrInternalServer, err.Error())
	}
	return nil
}

func DeleteMessageConsumer(ctx context.Context, messageUUID, handlerName string) *bmserror.BMSError {
	err := GetMessageOrm(ctx).Table(MessageConsumerTable).
		Where("message_uuid = ? and handler_name = ?", messageUUID, handlerName).Delete(&MMessageConsumer{}).GetError()
	if err != nil {
		return bmserror.NewError(constant.ErrInternalServer, err.Error())
	}
	return nil
}

func SetMessageStatus(ctx context.Context, id, toHandlerStatus int64) *bmserror.BMSError {
	db := GetMessageOrm(ctx).Table(MessageConsumerTable).Where("id = ?", id).
		Updates(map[string]interface{}{
			"handler_status": toHandlerStatus,
			"mtime":          time.Now().Unix(),
		})
	if err := db.GetError(); err != nil {
		return bmserror.NewError(constant.ErrInternalServer, err.Error())
	}
	return nil
}

func SetMessageStatusAndTimes(ctx context.Context, id, fromHandlerStatus, fromHandlerTimes, toHandlerStatus int64) *bmserror.BMSError {
	db := GetMessageOrm(ctx).Table(MessageConsumerTable).Where("id = ? and handler_status = ? and handler_times = ?", id, fromHandlerStatus, fromHandlerTimes).
		Updates(map[string]interface{}{
			"handler_status": toHandlerStatus,
			"handler_times":  fromHandlerTimes + 1,
			"mtime":          time.Now().Unix(),
		})
	if err := db.GetError(); err != nil {
		return bmserror.NewError(constant.ErrInternalServer, err.Error())
	}
	if db.RowsAffected() == 0 {
		return bmserror.NewError(constant.ErrInternalServer, "update status as ongoing fail|id:%v,fromHandlerStatus:%v,toHandlerStatus:%v,fromHandlerTimes:%v", id, fromHandlerStatus, toHandlerStatus, fromHandlerTimes)
	}
	return nil
}

func SetMessageConsumerHandlerStatus(ctx context.Context, messageUUID, handlerName string, handlerStatus int32) *bmserror.BMSError {
	m := &MMessageConsumer{}
	err := GetMessageOrm(ctx).Table(MessageConsumerTable).
		Where("message_uuid = ? and handler_name = ?", messageUUID, handlerName).Find(m).GetError()
	if err != nil {
		return bmserror.NewError(constant.ErrInternalServer, err.Error())
	}
	if m.ID == 0 {
		return nil
	}
	err = GetMessageOrm(ctx).Table(MessageConsumerTable).Where("id = ?", m.ID).
		Updates(map[string]interface{}{
			"handler_status": handlerStatus,
			"handler_times":  m.HandlerTimes + 1,
			"mtime":          time.Now().Unix(),
		}).GetError()
	if err != nil {
		return bmserror.NewError(constant.ErrInternalServer, err.Error())
	}
	return nil
}

func GetFromId(ctx context.Context) int64 {
	m := &MMessageConsumer{}
	// 获取10天前的 from id
	err := GetMessageOrm(ctx).Table(MessageConsumerTable).
		Where("ctime > ?", time.Now().Unix()-10*86400).Order("id").Limit(1).Find(&m).GetError()
	if err != nil {
		return 0
	}
	return m.ID
}

func GetProcessMessage(ctx context.Context, messageUUID string) (*MProcessMessage, *bmserror.BMSError) {
	l := make([]*MProcessMessage, 0)
	err := GetMessageOrm(ctx).Table(ProcessMessageTable).Where("message_uuid = ?", messageUUID).Find(&l).GetError()
	if err != nil {
		return nil, bmserror.NewError(constant.ErrInternalServer, err.Error())
	}
	if len(l) == 0 {
		return nil, nil
	}
	return l[0], nil
}

func GetFailedMessages(ctx context.Context, fromId, limit int64, retryTaskList []string, maxHandleTimes, thinkTaskTimeout int64) ([]*MRetryMessage, *bmserror.BMSError) {
	// 获取失败的任务，或者超时太久的任务
	now := time.Now().Unix()
	msgObjs := make([]*MRetryMessage, 0)
	db := GetMessageOrm(ctx).Replica().Table(MessageRetryTable).Select(MessageRetryColumn)
	db = db.Where("b.handler_status = ? or (b.handler_status in (?) and b.mtime < ?)", ConsumerHandlerFail, []int64{ConsumerHandlerOngoing, ConsumerHandlerPending}, now-thinkTaskTimeout)
	db = db.Where("b.id >= ? and a.task_name in (?) and b.handler_times < ?", fromId, retryTaskList, maxHandleTimes)
	db = db.Order("id").Limit(int(limit))
	err := db.Find(&msgObjs).GetError()
	if err != nil {
		return nil, bmserror.NewError(constant.ErrDB, err.Error())
	}
	return msgObjs, nil
}

type MessageUUIDRequest struct {
	MessageUUID string `gorm:"column:message_uuid" json:"message_uuid"`
}

// TODO:可以写得更优雅，将ManualProcessing放在AsyncRunner外面实现
func (r *AsyncRunner) ManualProcessing(s *web.BasicServer) {
	for taskName, processTask := range ProcessTaskMap {
		for funcName, tHandlerMethod := range processTask.HandlerMethod {
			handlerMethod := tHandlerMethod
			f := func(ctx context.Context, header *web.Header, request interface{}) (interface{}, *bmserror.BMSError) {
				err := handlerMethod(ctx, request)
				return nil, err
			}
			s.RegisterPOST("/api/v2/message/"+taskName+"/"+funcName, f, processTask.Message)
		}
	}

	for taskName, processTask := range ProcessTaskMap {
		for funcName, tHandlerMethod := range processTask.HandlerMethod {
			handlerMethod := tHandlerMethod
			messagePtr := processTask.Message
			tFuncName := funcName
			f := func(ctx context.Context, header *web.Header, request interface{}) (interface{}, *bmserror.BMSError) {
				req := request.(*MessageUUIDRequest)
				m, err := GetProcessMessage(ctx, req.MessageUUID)
				if err != nil {
					return nil, err
				}
				if m == nil {
					return nil, bmserror.NewError(constant.ErrInternalServer, "%s is not exist", req.MessageUUID)
				}

				t := reflect.TypeOf(messagePtr)
				if t.Kind() == reflect.Ptr {
					t = t.Elem()
				}
				message := reflect.New(t).Interface()

				errJson := json.Unmarshal([]byte(m.Message), message)
				if errJson != nil {
					return nil, bmserror.NewError(constant.ErrInternalServer, errJson.Error())
				}

				err = handlerMethod(ctx, message)
				if err == nil {
					_ = SetMessageConsumerHandlerStatus(ctx, req.MessageUUID, tFuncName, ConsumerHandlerSuccess)
				} else {
					_ = SetMessageConsumerHandlerStatus(ctx, req.MessageUUID, tFuncName, ConsumerHandlerFail)
				}

				return nil, err
			}
			s.RegisterPOST("/api/v2/message_uuid/"+taskName+"/"+funcName, f, &MessageUUIDRequest{})
		}
	}
}

func (r *AsyncRunner) InitRetryCronTask() {
	r.RegisterPayloadCrontabHandler("retry_task", r.DoHandleCronRetryInboundTask)
}

func (r *AsyncRunner) DoRetryOneTask(ctx context.Context, handlerMethod HandlerMethod, messagePtr interface{}, retryMsg *MRetryMessage) *bmserror.BMSError {
	t := reflect.TypeOf(messagePtr)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	message := reflect.New(t).Interface()

	errJson := json.Unmarshal([]byte(retryMsg.Message), message)
	if errJson != nil {
		_ = SetMessageStatus(ctx, retryMsg.ID, ConsumerHandlerFailTaskMsgDecodeError)
		return bmserror.NewError(constant.ErrInternalServer, errJson.Error())
	}

	err := handlerMethod(ctx, message)
	if err != nil {
		// 需要继续重试
		_ = SetMessageStatus(ctx, retryMsg.ID, ConsumerHandlerFail)
		return err.Mark()
	}
	_ = SetMessageStatus(ctx, retryMsg.ID, ConsumerHandlerSuccess)
	return nil
}

func (r *AsyncRunner) RetryOneTask(ctx context.Context, retryMsg *MRetryMessage, maxHandleTimes int32) (isSkip bool) {
	isSkip = false
	defer func() {
		if panicErr := recover(); panicErr != nil {
			exception := string(debug.Stack())
			log.Errorf("retry_task panic error:%v, taskName:%v, message:%v, uuid:%v|exception:%v", panicErr, retryMsg.TaskName, retryMsg.Message, retryMsg.MessageUUID, exception)
			_ = monitor.AwesomeReportEventWithoutTrans(ctx, CronTaskModule, retryMsg.TaskName, "-1", exception)
			_ = SetMessageStatus(ctx, retryMsg.ID, ConsumerHandlerException)
		}
	}()
	if retryMsg.HandlerTimes == maxHandleTimes-1 {
		_ = monitor.AwesomeReportEventWithoutTrans(ctx, CronTaskModule, retryMsg.TaskName, "-1", "exceed max retry")
	}
	// 更新为ongoing
	if err := SetMessageStatusAndTimes(ctx, retryMsg.ID, int64(retryMsg.HandlerStatus), int64(retryMsg.HandlerTimes), ConsumerHandlerOngoing); err != nil {
		if err.Code() == constant.ErrInternalServer {
			isSkip = true
		}
		return isSkip
	}
	task, ok := ProcessTaskMap[retryMsg.TaskName]
	if !ok {
		// 系统没有注册对应的消息处理方法
		_ = SetMessageStatus(ctx, retryMsg.ID, ConsumerHandlerFailTaskNotFound)
		_ = monitor.AwesomeReportEventWithoutTrans(ctx, CronTaskModule, retryMsg.TaskName, "-1", "task not found")
		return isSkip
	}
	// 如果方法名被修改，将导致任务无法重试，所以如果发现只注册了一个方法的话，认为是跟之前相同的任务
	var handlerMethod HandlerMethod
	if len(task.HandlerMethod) == 1 {
		for _, tHandlerMethod := range task.HandlerMethod {
			handlerMethod = tHandlerMethod
		}
	} else {
		handlerMethod, ok = task.HandlerMethod[retryMsg.HandlerName]
		if !ok {
			// 系统没有注册对应的消息处理方法
			_ = SetMessageStatus(ctx, retryMsg.ID, ConsumerHandlerFailTaskNotFound)
			_ = monitor.AwesomeReportEventWithoutTrans(ctx, CronTaskModule, retryMsg.TaskName, "-1", "task not found")
			return isSkip
		}
	}

	err := r.DoRetryOneTask(ctx, handlerMethod, task.Message, retryMsg)
	if err != nil {
		catData := fmt.Sprintf("uuid:%s,error:%s", retryMsg.MessageUUID, err.DebugError())
		_ = monitor.AwesomeReportEventWithoutTrans(ctx, CronTaskModule, retryMsg.TaskName, "-1", catData)
		return isSkip
	}
	return isSkip
}

func (r *AsyncRunner) RetryTask(payload *RetryTaskPayLoad) *bmserror.BMSError {
	ctx := context.TODO()
	if gFromId == 0 {
		gFromId = GetFromId(ctx)
	}
	fromId := gFromId
	for idx := 0; idx < payload.Count; idx++ {
		msgObjs, err := GetFailedMessages(ctx, fromId, payload.Limit, payload.TaskList, payload.MaxHandleTimes, payload.ThinkTaskTimeout)
		if err != nil {
			log.Errorf("get failed message fail|error:%v", err)
			return err.Mark()
		}
		for _, msgObj := range msgObjs {
			if isSkip := r.RetryOneTask(ctx, msgObj, int32(payload.MaxHandleTimes)); isSkip {
				// 可能有并发的定时任务，退出本地处理
				break
			}
		}
		if idx == payload.Count-1 {
			_ = monitor.AwesomeReportEventWithoutTrans(ctx, CronTaskModule, "too_many_retry_task", "-1", "too many, please check")
		}
		// 开始处理任务
		if len(msgObjs) < int(payload.Limit) {
			break
		}
		fromId = msgObjs[len(msgObjs)-1].ID
	}
	return nil
}

type RetryTaskPayLoad struct {
	TaskList         []string `json:"task_list"`           // 任务列表
	Count            int      `json:"count"`               // 循环多少次
	Limit            int64    `json:"limit"`               // 每个循环最多取多少条数据
	MaxHandleTimes   int64    `json:"max_handle_times"`    // 单个任务最多重试多少次
	ThinkTaskTimeout int64    `json:"think_task_time_out"` // 多少秒任务任务处理超时了，需要重试，最小必须大于120s
}

const CronTaskModule = "CronTask"

func (r *AsyncRunner) DoHandleCronRetryInboundTask(ctx context.Context, payload *crontask.TaskPayload) *bmserror.BMSError {
	var err *bmserror.BMSError
	defer func() {
		if err != nil {
			_ = monitor.AwesomeReportEventWithoutTrans(ctx, CronTaskModule, "Retry Task Param Error", "-1", payload.Param)
		}
	}()
	retryTaskPayLoad := &RetryTaskPayLoad{}
	jsonErr := json.Unmarshal([]byte(payload.Param), &retryTaskPayLoad)
	if jsonErr != nil {
		err = bmserror.NewError(constant.ErrJsonEncodeFail, jsonErr.Error())
		return err
	}
	if len(retryTaskPayLoad.TaskList) == 0 {
		err = bmserror.NewError(constant.ErrParam, "TaskList should not empty")
		return err
	}
	if retryTaskPayLoad.Count <= 0 {
		err = bmserror.NewError(constant.ErrParam, "Count should > 0")
		return err
	}
	if retryTaskPayLoad.Limit <= 0 {
		err = bmserror.NewError(constant.ErrParam, "Limit should > 0")
		return err
	}
	if retryTaskPayLoad.MaxHandleTimes <= 1 {
		err = bmserror.NewError(constant.ErrParam, "MaxHandleTimes should > 1")
		return err
	}
	if retryTaskPayLoad.ThinkTaskTimeout < 120 {
		err = bmserror.NewError(constant.ErrParam, "ThinkTaskTimeout should >= 120")
		return err
	}
	//获取当前仓库配置的所有连接之后，直接按照对应仓库的ctx进行查询和处理
	var errs []*bmserror.BMSError
	err = r.RetryTask(retryTaskPayLoad)
	if err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return bmserror.FormatErrs(errs)
	}
	return nil
}
