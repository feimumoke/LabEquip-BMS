package manager

import (
	"context"

	"github.com/feimumoke/labequipbms/defines/constant"
	"github.com/feimumoke/labequipbms/defines/entity"
	"github.com/feimumoke/labequipbms/framework/bmserror"
	"github.com/feimumoke/labequipbms/framework/datasource"
	"github.com/feimumoke/labequipbms/framework/support/paginator"
	"github.com/feimumoke/labequipbms/framework/support/timeutil"
)

type BorrowTaskManager struct {
	ds datasource.DataSource
}

func NewBorrowTaskManager() *BorrowTaskManager {
	return &BorrowTaskManager{ds: datasource.DefaultBMSSource}
}

// CreateBorrowTask 创建借记任务
func (m *BorrowTaskManager) CreateBorrowTask(ctx context.Context, task *entity.BorrowTask) *bmserror.BMSError {
	if err := m.ds.GetDataSource(ctx, nil).Table(entity.BorrowTaskTabName).Create(task).GetError(); err != nil {
		return bmserror.NewError(constant.ErrDB, err.Error())
	}
	return nil
}

// CreateBorrowTaskLog 创建借记任务日志
func (m *BorrowTaskManager) CreateBorrowTaskLog(ctx context.Context, log *entity.BorrowTaskLog) *bmserror.BMSError {
	if err := m.ds.GetDataSource(ctx, nil).Table(entity.BorrowTaskLogTabName).Create(log).GetError(); err != nil {
		return bmserror.NewError(constant.ErrDB, err.Error())
	}
	return nil
}

// GetBorrowTaskByTaskId 根据任务ID获取借记任务
func (m *BorrowTaskManager) GetBorrowTaskByTaskId(ctx context.Context, taskId string, withLock bool) (*entity.BorrowTask, *bmserror.BMSError) {
	var task entity.BorrowTask
	db := m.ds.GetDataSource(ctx, nil).Table(entity.BorrowTaskTabName).
		Where("task_id = ?", taskId)
	if withLock {
		db = db.ForUpdate()
	}
	if err := db.First(&task).GetError(); err != nil {
		if db.RecordNotFound() {
			return nil, nil
		}
		return nil, bmserror.NewError(constant.ErrDB, err.Error())
	}
	return &task, nil
}

// UpdateBorrowTask 更新借记任务
func (m *BorrowTaskManager) UpdateBorrowTask(ctx context.Context, task *entity.BorrowTask) *bmserror.BMSError {
	now := timeutil.GetCurrentUnix()
	updateMap := map[string]interface{}{
		"task_status": task.TaskStatus,
		"approval":    task.Approval,
		"mtime":       now,
	}
	if err := m.ds.GetDataSource(ctx, nil).Table(entity.BorrowTaskTabName).
		Where("task_id = ?", task.TaskID).
		Updates(updateMap).GetError(); err != nil {
		return bmserror.NewError(constant.ErrDB, err.Error())
	}
	task.Mtime = now
	return nil
}

// SearchBorrowTask 查询借记任务
type BorrowTaskSearchParam struct {
	LabCode    string
	EquipId    string
	Operator   string
	TaskStatus constant.BorrowTaskStatus
	PageIn     *paginator.PageIn
}

func (m *BorrowTaskManager) SearchBorrowTask(ctx context.Context, params *BorrowTaskSearchParam) ([]*entity.BorrowTask, int64, *bmserror.BMSError) {
	var taskList []*entity.BorrowTask
	db := m.ds.GetDataSource(ctx, nil).Table(entity.BorrowTaskTabName)
	if params.LabCode != "" {
		db = db.Where("lab_id = ?", params.LabCode)
	}
	if params.EquipId != "" {
		db = db.Where("equip_id = ?", params.EquipId)
	}
	if params.Operator != "" {
		db = db.Where("creator = ?", params.Operator)
	}
	if params.TaskStatus > 0 {
		db = db.Where("task_status = ?", params.TaskStatus)
	}
	total, err := paginator.Paginator(db, params.PageIn, &taskList)
	if err != nil {
		return nil, 0, err.Mark()
	}
	return taskList, total, nil
}

// GetBorrowTaskLogs 获取借记任务日志
func (m *BorrowTaskManager) GetBorrowTaskLogs(ctx context.Context, taskId string) ([]*entity.BorrowTaskLog, *bmserror.BMSError) {
	var logs []*entity.BorrowTaskLog
	if err := m.ds.GetDataSource(ctx, nil).Table(entity.BorrowTaskLogTabName).
		Where("task_id = ?", taskId).
		Order("id ASC").
		Find(&logs).GetError(); err != nil {
		return nil, bmserror.NewError(constant.ErrDB, err.Error())
	}
	return logs, nil
}

// BatchGetBorrowTaskLogs 批量获取借记任务日志
func (m *BorrowTaskManager) BatchGetBorrowTaskLogs(ctx context.Context, taskIdList []string) (map[string][]*entity.BorrowTaskLog, *bmserror.BMSError) {
	if len(taskIdList) == 0 {
		return make(map[string][]*entity.BorrowTaskLog), nil
	}
	var logs []*entity.BorrowTaskLog
	if err := m.ds.GetDataSource(ctx, nil).Table(entity.BorrowTaskLogTabName).
		Where("task_id IN (?)", taskIdList).
		Order("task_id ASC, id ASC").
		Find(&logs).GetError(); err != nil {
		return nil, bmserror.NewError(constant.ErrDB, err.Error())
	}

	logMap := make(map[string][]*entity.BorrowTaskLog)
	for _, log := range logs {
		logMap[log.TaskID] = append(logMap[log.TaskID], log)
	}
	return logMap, nil
}
