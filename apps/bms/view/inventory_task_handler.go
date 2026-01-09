package view

import (
	"context"

	pbbms "github.com/feimumoke/labequipbms/api_idl/apps/bms"
	"github.com/feimumoke/labequipbms/apps/basic/manager"
	bmsmanager "github.com/feimumoke/labequipbms/apps/bms/manager"
	"github.com/feimumoke/labequipbms/defines/entity"
	"github.com/feimumoke/labequipbms/framework/bmserror"
	"github.com/feimumoke/labequipbms/framework/support/convert"
	"github.com/feimumoke/labequipbms/framework/support/paginator"
	"github.com/feimumoke/labequipbms/framework/web"
)

type InventoryTaskHandler struct {
	inventoryTaskMng *bmsmanager.InventoryTaskManager
	inventoryMng     *bmsmanager.InventoryManager
	equipMng         *manager.EquipManager
	labMng           *manager.LabManager
}

func NewInventoryTaskHandler() *InventoryTaskHandler {
	return &InventoryTaskHandler{
		inventoryTaskMng: bmsmanager.NewInventoryTaskManager(),
		inventoryMng:     bmsmanager.NewInventoryManager(),
		equipMng:         manager.NewEquipManager(),
		labMng:           manager.NewLabManager(),
	}
}

// SearchInvTaskHandler 库存任务查询接口
func (h *InventoryTaskHandler) SearchInvTaskHandler(ctx context.Context, header *web.Header, i interface{}) (interface{}, *bmserror.BMSError) {
	req := i.(*pbbms.SearchInvTaskRequest)
	pageIn := &paginator.PageIn{
		Pageno:     req.GetPageno(),
		Count:      req.GetCount(),
		IsGetTotal: true,
	}

	taskList, total, bmsErr := h.inventoryTaskMng.SearchInventoryTask(ctx, &bmsmanager.InventoryTaskSearchParam{
		LabCode: req.GetLabCode(),
		EquipId: req.GetEquipId(),
		PageIn:  pageIn,
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

	// 获取实验室信息
	labMap := make(map[string]*entity.LaboratoryTab)
	labList, _ := h.labMng.GetAllLabMng(ctx)
	for _, lab := range labList {
		labMap[lab.LabCode] = lab
	}

	// 获取库存信息
	inventoryMap := make(map[string]*entity.InventoryTab)
	for _, task := range taskList {
		key := task.LabId + "_" + task.EquipId
		if _, ok := inventoryMap[key]; !ok {
			inventory, _ := h.inventoryMng.GetInventoryByLabAndEquip(ctx, task.LabId, task.EquipId, false)
			if inventory != nil {
				inventoryMap[key] = inventory
			}
		}
	}

	invTaskInfoList := make([]*pbbms.InvTaskInfo, 0, len(taskList))
	for _, task := range taskList {
		equip := equipMap[task.EquipId]
		lab := labMap[task.LabId]
		key := task.LabId + "_" + task.EquipId
		inventory := inventoryMap[key]

		equipName := ""
		if equip != nil {
			equipName = equip.EquipName
		}
		labName := ""
		if lab != nil {
			labName = lab.LabName
		}

		totalQty := int64(0)
		availableQty := int64(0)
		borrowedQty := int64(0)
		preAllocatedQty := int64(0)
		reservedQty := int64(0)
		if inventory != nil {
			totalQty = inventory.TotalQty
			availableQty = inventory.AvailableQty
			borrowedQty = inventory.BorrowedQty
			preAllocatedQty = inventory.AllocatedQty
		}

		// 获取任务日志
		logs, _ := h.inventoryTaskMng.GetInventoryTaskLogs(ctx, task.TaskID)
		logList := make([]*pbbms.InvTaskLog, 0, len(logs))
		for _, log := range logs {
			logList = append(logList, &pbbms.InvTaskLog{
				TaskStatus: convert.Int64(log.TaskStatus),
				Operator:   &log.Operator,
				Message:    &log.Remark,
				Info:       &log.Remark,
				Ctime:      convert.Int64(0), // InventoryTaskLogTab中没有Ctime字段
			})
		}

		invTaskInfo := &pbbms.InvTaskInfo{
			TaskId:          &task.TaskID,
			EquipId:         &task.EquipId,
			EquipName:       []string{equipName},
			LabCode:         &task.LabId,
			LabName:         &labName,
			TotalQty:        &totalQty,
			AvailableQty:    &availableQty,
			BorrowedQty:     &borrowedQty,
			PreAllocatedQty: &preAllocatedQty,
			ReservedQty:     &reservedQty,
			LogList:         logList,
		}
		invTaskInfoList = append(invTaskInfoList, invTaskInfo)
	}

	return &pbbms.SearchInvTaskResponse{
		Total: convert.Int64(total),
		List:  invTaskInfoList,
	}, nil
}
