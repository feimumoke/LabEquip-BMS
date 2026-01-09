package manager

import (
	"context"

	"github.com/feimumoke/labequipbms/defines/constant"
	"github.com/feimumoke/labequipbms/defines/entity"
	"github.com/feimumoke/labequipbms/framework/bmserror"
	"github.com/feimumoke/labequipbms/framework/datasource"
	"github.com/feimumoke/labequipbms/framework/support/paginator"
)

type InventoryTaskManager struct {
	ds datasource.DataSource
}

func NewInventoryTaskManager() *InventoryTaskManager {
	return &InventoryTaskManager{ds: datasource.DefaultInvSource}
}

// CreateInventoryTask 创建库存任务
func (m *InventoryTaskManager) CreateInventoryTask(ctx context.Context, task *entity.InventoryTaskTab) *bmserror.BMSError {
	if err := m.ds.GetDataSource(ctx, nil).Table(entity.InventoryTaskTabName).Create(task).GetError(); err != nil {
		return bmserror.NewError(constant.ErrDB, err.Error())
	}
	return nil
}

// CreateInventoryTaskLog 创建库存任务日志
func (m *InventoryTaskManager) CreateInventoryTaskLog(ctx context.Context, log *entity.InventoryTaskLogTab) *bmserror.BMSError {
	if err := m.ds.GetDataSource(ctx, nil).Table(entity.InventoryTaskLogTabName).Create(log).GetError(); err != nil {
		return bmserror.NewError(constant.ErrDB, err.Error())
	}
	return nil
}

// SearchInventoryTask 查询库存任务
type InventoryTaskSearchParam struct {
	LabCode string
	EquipId string
	PageIn  *paginator.PageIn
}

func (m *InventoryTaskManager) SearchInventoryTask(ctx context.Context, params *InventoryTaskSearchParam) ([]*entity.InventoryTaskTab, int64, *bmserror.BMSError) {
	var taskList []*entity.InventoryTaskTab
	db := m.ds.GetDataSource(ctx, nil).Table(entity.InventoryTaskTabName)
	if params.LabCode != "" {
		db = db.Where("lab_id = ?", params.LabCode)
	}
	if params.EquipId != "" {
		db = db.Where("equip_id = ?", params.EquipId)
	}
	total, err := paginator.Paginator(db, params.PageIn, &taskList)
	if err != nil {
		return nil, 0, err.Mark()
	}
	return taskList, total, nil
}

// GetInventoryTaskLogs 获取库存任务日志
func (m *InventoryTaskManager) GetInventoryTaskLogs(ctx context.Context, taskId string) ([]*entity.InventoryTaskLogTab, *bmserror.BMSError) {
	var logs []*entity.InventoryTaskLogTab
	if err := m.ds.GetDataSource(ctx, nil).Table(entity.InventoryTaskLogTabName).
		Where("task_id = ?", taskId).
		Order("id ASC").
		Find(&logs).GetError(); err != nil {
		return nil, bmserror.NewError(constant.ErrDB, err.Error())
	}
	return logs, nil
}

// BatchGetInventoryTaskLogs 批量获取库存任务日志
func (m *InventoryTaskManager) BatchGetInventoryTaskLogs(ctx context.Context, taskIdList []string) (map[string][]*entity.InventoryTaskLogTab, *bmserror.BMSError) {
	if len(taskIdList) == 0 {
		return make(map[string][]*entity.InventoryTaskLogTab), nil
	}
	var logs []*entity.InventoryTaskLogTab
	if err := m.ds.GetDataSource(ctx, nil).Table(entity.InventoryTaskLogTabName).
		Where("task_id IN (?)", taskIdList).
		Order("task_id ASC, id ASC").
		Find(&logs).GetError(); err != nil {
		return nil, bmserror.NewError(constant.ErrDB, err.Error())
	}

	logMap := make(map[string][]*entity.InventoryTaskLogTab)
	for _, log := range logs {
		logMap[log.TaskID] = append(logMap[log.TaskID], log)
	}
	return logMap, nil
}

// GetInventoryTaskByTaskId 根据任务ID获取库存任务
func (m *InventoryTaskManager) GetInventoryTaskByTaskId(ctx context.Context, taskId string) (*entity.InventoryTaskTab, *bmserror.BMSError) {
	var task entity.InventoryTaskTab
	db := m.ds.GetDataSource(ctx, nil).Table(entity.InventoryTaskTabName).
		Where("task_id = ?", taskId)
	if err := db.First(&task).GetError(); err != nil {
		if db.RecordNotFound() {
			return nil, nil
		}
		return nil, bmserror.NewError(constant.ErrDB, err.Error())
	}
	return &task, nil
}
