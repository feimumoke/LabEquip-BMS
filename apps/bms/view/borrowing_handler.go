package view

import (
	"context"

	pbbms "github.com/feimumoke/labequipbms/api_idl/apps/bms"
	"github.com/feimumoke/labequipbms/apps/basic/manager"
	bmsmanager "github.com/feimumoke/labequipbms/apps/bms/manager"
	bmservice "github.com/feimumoke/labequipbms/apps/bms/service"
	"github.com/feimumoke/labequipbms/defines/constant"
	"github.com/feimumoke/labequipbms/defines/entity"
	"github.com/feimumoke/labequipbms/framework/bmserror"
	"github.com/feimumoke/labequipbms/framework/support/convert"
	"github.com/feimumoke/labequipbms/framework/support/paginator"
	"github.com/feimumoke/labequipbms/framework/web"
)

type BorrowingHandler struct {
	inventoryService *bmservice.InventoryService
	borrowService    *bmservice.BorrowService
	borrowTaskMng    *bmsmanager.BorrowTaskManager
	equipMng         *manager.EquipManager
	labMng           *manager.LabManager
}

func NewBorrowingHandler() *BorrowingHandler {
	return &BorrowingHandler{
		inventoryService: bmservice.NewInventoryService(),
		borrowService:    bmservice.NewBorrowService(),
		borrowTaskMng:    bmsmanager.NewBorrowTaskManager(),
		equipMng:         manager.NewEquipManager(),
		labMng:           manager.NewLabManager(),
	}
}

// CreateBorrowHandler 借记任务创建接口
func (h *BorrowingHandler) CreateBorrowHandler(ctx context.Context, header *web.Header, i interface{}) (interface{}, *bmserror.BMSError) {
	req := i.(*pbbms.CreateBorrowRequest)
	if req.GetEquipId() == "" {
		return nil, bmserror.NewError(constant.ErrParam, "equip_id is empty")
	}
	if req.GetLabCode() == "" {
		return nil, bmserror.NewError(constant.ErrParam, "lab_id is empty")
	}
	if req.GetBorrowQty() <= 0 {
		return nil, bmserror.NewError(constant.ErrParam, "count must be greater than 0")
	}

	// 获取实验室信息
	lab, bmsErr := h.labMng.GetLabByCode(ctx, req.GetLabCode())
	if bmsErr != nil {
		return nil, bmsErr.Mark()
	}
	if lab == nil {
		return nil, bmserror.NewError(constant.ErrParam, "lab not found")
	}

	// 调用 service 创建借记任务
	_, _, bmsErr = h.borrowService.CreateBorrowTask(ctx, lab.LabCode, req.GetEquipId(), req.GetBorrowQty(), req.GetDescription(), header.UserEmail)
	if bmsErr != nil {
		return nil, bmsErr.Mark()
	}

	return &pbbms.CreateBorrowResponse{}, nil
}

// CancelBorrowHandler 借记任务取消接口
func (h *BorrowingHandler) CancelBorrowHandler(ctx context.Context, header *web.Header, i interface{}) (interface{}, *bmserror.BMSError) {
	req := i.(*pbbms.CancelBorrowRequest)
	if req.GetBorrowId() == "" {
		return nil, bmserror.NewError(constant.ErrParam, "borrow_id is empty")
	}

	// 调用 service 取消借记任务
	bmsErr := h.borrowService.CancelBorrowTask(ctx, req.GetBorrowId(), req.GetReason(), header.UserEmail)
	if bmsErr != nil {
		return nil, bmsErr.Mark()
	}

	return &pbbms.CancelBorrowResponse{}, nil
}

// TaskBorrowHandler 拿走借记物品接口
func (h *BorrowingHandler) TaskBorrowHandler(ctx context.Context, header *web.Header, i interface{}) (interface{}, *bmserror.BMSError) {
	req := i.(*pbbms.TaskBorrowRequest)
	if req.GetBorrowId() == "" {
		return nil, bmserror.NewError(constant.ErrParam, "borrow_id is empty")
	}

	// 调用 service 取消借记任务
	bmsErr := h.borrowService.TaskBorrowTask(ctx, req.GetBorrowId(), header.UserEmail, req.GetCodeList())
	if bmsErr != nil {
		return nil, bmsErr.Mark()
	}

	return &pbbms.TakeBorrowResponse{}, nil
}

// ReturnBorrowHandler 归还接口
func (h *BorrowingHandler) ReturnBorrowHandler(ctx context.Context, header *web.Header, i interface{}) (interface{}, *bmserror.BMSError) {
	req := i.(*pbbms.ReturnBorrowRequest)
	if req.GetBorrowId() == "" {
		return nil, bmserror.NewError(constant.ErrParam, "borrow_id is empty")
	}
	if req.GetReturnQty() <= 0 {
		return nil, bmserror.NewError(constant.ErrParam, "return qty must be greater than 0")
	}

	// 调用 service 完成借记任务
	bmsErr := h.borrowService.CompleteBorrowTask(ctx, req.GetBorrowId(), header.UserEmail, req.GetReturnQty())
	if bmsErr != nil {
		return nil, bmsErr.Mark()
	}

	return &pbbms.ReturnBorrowResponse{}, nil
}

// SearchBorrowTaskHandler 借记任务查询接口
func (h *BorrowingHandler) SearchBorrowTaskHandler(ctx context.Context, header *web.Header, i interface{}) (interface{}, *bmserror.BMSError) {
	req := i.(*pbbms.SearchBorrowTaskRequest)
	pageIn := &paginator.PageIn{
		Pageno:     req.GetPageno(),
		Count:      req.GetCount(),
		IsGetTotal: true,
	}

	var taskStatus constant.BorrowTaskStatus
	if req.GetTaskStatus() > 0 {
		taskStatus = req.GetTaskStatus()
	}

	taskList, total, bmsErr := h.borrowTaskMng.SearchBorrowTask(ctx, &bmsmanager.BorrowTaskSearchParam{
		LabCode:    req.GetLabCode(),
		EquipId:    req.GetEquipId(),
		Operator:   req.GetOperator(),
		TaskStatus: taskStatus,
		PageIn:     pageIn,
	})
	if bmsErr != nil {
		return nil, bmsErr.Mark()
	}

	// 获取设备信息
	equipMap := make(map[string]*entity.EquipTab)
	if len(taskList) > 0 {
		equipIdList := make([]string, 0, len(taskList))
		for _, task := range taskList {
			equipIdList = append(equipIdList, task.EquipId)
		}
		equipList, _, _ := h.equipMng.SearchEquipMng(ctx, &manager.EquipSearchParam{
			EquipIdList: equipIdList,
			PageIn:      nil,
		})
		for _, equip := range equipList {
			equipMap[equip.EquipId] = equip
		}
	}

	// 批量获取实验室信息
	labCodeSet := make(map[string]bool)
	for _, task := range taskList {
		labCodeSet[task.LabId] = true
	}
	labCodeList := make([]string, 0, len(labCodeSet))
	for labCode := range labCodeSet {
		labCodeList = append(labCodeList, labCode)
	}
	labMap, bmsErr := h.labMng.BatchGetLabByCodes(ctx, labCodeList)
	if bmsErr != nil {
		return nil, bmsErr.Mark()
	}

	// 批量获取任务日志
	taskIdList := make([]string, 0, len(taskList))
	for _, task := range taskList {
		taskIdList = append(taskIdList, task.TaskID)
	}
	logMap, bmsErr := h.borrowTaskMng.BatchGetBorrowTaskLogs(ctx, taskIdList)
	if bmsErr != nil {
		return nil, bmsErr.Mark()
	}

	borrowTaskInfoList := make([]*pbbms.BorrowTaskInfo, 0, len(taskList))
	for _, task := range taskList {
		equip := equipMap[task.EquipId]
		lab := labMap[task.LabId]

		equipName := ""
		if equip != nil {
			equipName = equip.EquipName
		}
		labName := ""
		if lab != nil {
			labName = lab.LabName
		}

		// 从批量查询的日志中获取
		logs := logMap[task.TaskID]
		logList := make([]*pbbms.BorrowTaskLog, 0, len(logs))
		for _, log := range logs {
			logList = append(logList, &pbbms.BorrowTaskLog{
				TaskStatus: convert.Int64(log.TaskStatus),
				Operator:   &log.Operator,
				Message:    &log.Remark,
				Info:       &log.Remark,
				Ctime:      convert.Int64(log.Ctime),
			})
		}

		borrowQty := task.BorrowQty
		borrowTaskInfo := &pbbms.BorrowTaskInfo{
			TaskId:     &task.TaskID,
			EquipId:    &task.EquipId,
			EquipName:  convert.String(equipName),
			LabCode:    &task.LabId,
			LabName:    &labName,
			BorrowQty:  &borrowQty,
			TaskStatus: convert.Int64(task.TaskStatus),
			Operator:   &task.Creator,
			Ctime:      convert.Int64(task.Ctime),
			LogList:    logList,
		}
		borrowTaskInfoList = append(borrowTaskInfoList, borrowTaskInfo)
	}

	return &pbbms.SearchBorrowTaskResponse{
		Total: convert.Int64(total),
		List:  borrowTaskInfoList,
	}, nil
}
