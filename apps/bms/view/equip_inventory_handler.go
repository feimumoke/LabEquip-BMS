package view

import (
	"context"

	pbbms "github.com/feimumoke/labequipbms/api_idl/apps/bms"
	"github.com/feimumoke/labequipbms/apps/basic/manager"
	bmsmanager "github.com/feimumoke/labequipbms/apps/bms/manager"
	bmservice "github.com/feimumoke/labequipbms/apps/bms/service"
	"github.com/feimumoke/labequipbms/apps/common/idutil"
	"github.com/feimumoke/labequipbms/defines/constant"
	"github.com/feimumoke/labequipbms/defines/entity"
	"github.com/feimumoke/labequipbms/framework/bmserror"
	"github.com/feimumoke/labequipbms/framework/support/convert"
	"github.com/feimumoke/labequipbms/framework/support/paginator"
	"github.com/feimumoke/labequipbms/framework/support/timeutil"
	"github.com/feimumoke/labequipbms/framework/transaction"
	"github.com/feimumoke/labequipbms/framework/web"
)

type EquipInventoryHandler struct {
	inventoryService *bmservice.InventoryService
	inventoryMng     *bmsmanager.InventoryManager
	inventoryTaskMng *bmsmanager.InventoryTaskManager
	equipMng         *manager.EquipManager
	labMng           *manager.LabManager
}

func NewEquipInventoryHandler() *EquipInventoryHandler {
	return &EquipInventoryHandler{
		inventoryService: bmservice.NewInventoryService(),
		inventoryMng:     bmsmanager.NewInventoryManager(),
		inventoryTaskMng: bmsmanager.NewInventoryTaskManager(),
		equipMng:         manager.NewEquipManager(),
		labMng:           manager.NewLabManager(),
	}
}

// CreateEquipInvHandler 增加设备库存接口
func (h *EquipInventoryHandler) CreateEquipInvHandler(ctx context.Context, header *web.Header, i interface{}) (interface{}, *bmserror.BMSError) {
	req := i.(*pbbms.CreateEquipInvRequest)
	if req.GetEquipId() == "" {
		return nil, bmserror.NewError(constant.ErrParam, "equip_id is empty")
	}
	if req.GetLabCode() == "" {
		return nil, bmserror.NewError(constant.ErrParam, "lab_code is empty")
	}
	if req.GetCount() <= 0 {
		return nil, bmserror.NewError(constant.ErrParam, "count must be greater than 0")
	}

	// 校验设备是否存在
	equipList, _, bmsErr := h.equipMng.SearchEquipMng(ctx, &manager.EquipSearchParam{
		EquipIdList: []string{req.GetEquipId()},
		PageIn:      nil,
	})
	if bmsErr != nil {
		return nil, bmsErr.Mark()
	}
	if len(equipList) == 0 {
		return nil, bmserror.NewError(constant.ErrParam, "equip not found")
	}

	// 获取实验室信息
	lab, bmsErr := h.labMng.GetLabByCode(ctx, req.GetLabCode())
	if bmsErr != nil {
		return nil, bmsErr.Mark()
	}
	if lab == nil {
		return nil, bmserror.NewError(constant.ErrParam, "lab not found")
	}

	// 生成任务ID
	taskId, bmsErr := idutil.GenerateTaskNumber(ctx, constant.EquipInvTaskId)
	if bmsErr != nil {
		return nil, bmsErr.Mark()
	}

	// 调用 service 增加库存和记录三级账

	tErr := transaction.PropagationRequired(ctx, func(ctx context.Context) *bmserror.BMSError {
		stockReq := &bmservice.IncreaseStockRequest{
			LabId:          req.GetLabCode(),
			EquipId:        req.GetEquipId(),
			Count:          req.GetCount(),
			SheetId:        taskId,
			TransSheetType: constant.TransactionSheetTypeInventory,
			Operator:       header.UserEmail,
			Description:    req.GetDescription(),
		}
		_, bmsErr = h.inventoryService.IncreaseInventory(ctx, stockReq)
		if bmsErr != nil {
			return bmsErr.Mark()
		}
		// 创建库存任务
		now := timeutil.GetCurrentUnix()
		task := &entity.InventoryTaskTab{
			TaskID:     taskId,
			TaskType:   constant.InventoryTaskTypeIncrease,
			TaskStatus: constant.InventoryTaskStatusDone,
			LabId:      lab.LabCode,
			EquipId:    req.GetEquipId(),
			TotalQty:   req.GetCount(),
			Operator:   header.UserEmail,
			Remark:     req.GetDescription(),
			Ctime:      now,
			Mtime:      now,
		}
		if bmsErr := h.inventoryTaskMng.CreateInventoryTask(ctx, task); bmsErr != nil {
			return bmsErr.Mark()
		}
		// 创建任务日志
		taskLog := &entity.InventoryTaskLogTab{
			TaskID:     taskId,
			TaskStatus: constant.InventoryTaskStatusDone,
			Remark:     req.GetDescription(),
			Operator:   header.UserEmail,
		}
		if bmsErr := h.inventoryTaskMng.CreateInventoryTaskLog(ctx, taskLog); bmsErr != nil {
			return bmsErr.Mark()
		}
		return nil

	})
	if tErr != nil {
		return nil, tErr.Mark()
	}
	return &pbbms.CreateEquipInvResponse{}, nil
}

// SearchEquipInvHandler 库存查询接口
func (h *EquipInventoryHandler) SearchEquipInvHandler(ctx context.Context, header *web.Header, i interface{}) (interface{}, *bmserror.BMSError) {
	req := i.(*pbbms.SearchEquipRequest)
	pageIn := &paginator.PageIn{
		Pageno:     1,
		Count:      100,
		IsGetTotal: true,
	}
	inventoryList, total, bmsErr := h.inventoryMng.SearchInventory(ctx, &bmsmanager.InventorySearchParam{
		LabCode: req.GetLabCode(),
		EquipId: req.GetEquipId(),
		PageIn:  pageIn,
	})
	if bmsErr != nil {
		return nil, bmsErr.Mark()
	}

	// 获取设备信息
	equipMap := make(map[string]*entity.EquipTab)
	if len(inventoryList) > 0 {
		equipIdList := make([]string, 0, len(inventoryList))
		for _, inv := range inventoryList {
			equipIdList = append(equipIdList, inv.EquipId)
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
	for _, inv := range inventoryList {
		labCodeSet[inv.LabId] = true
	}
	labCodeList := make([]string, 0, len(labCodeSet))
	for labCode := range labCodeSet {
		labCodeList = append(labCodeList, labCode)
	}
	labMap, bmsErr := h.labMng.BatchGetLabByCodes(ctx, labCodeList)
	if bmsErr != nil {
		return nil, bmsErr.Mark()
	}

	equipInvInfoList := make([]*pbbms.EquipInvInfo, 0, len(inventoryList))
	for _, inv := range inventoryList {
		equip := equipMap[inv.EquipId]
		lab := labMap[inv.LabId]

		equipName := ""
		if equip != nil {
			equipName = equip.EquipName
		}
		labName := ""
		if lab != nil {
			labName = lab.LabName
		}

		equipInvInfo := &pbbms.EquipInvInfo{
			EquipId:         &inv.EquipId,
			EquipName:       []string{equipName},
			LabCode:         &inv.LabId,
			LabName:         &labName,
			TotalQty:        &inv.TotalQty,
			AvailableQty:    &inv.AvailableQty,
			BorrowedQty:     &inv.BorrowedQty,
			PreAllocatedQty: &inv.AllocatedQty,
			ReservedQty:     nil,
		}
		equipInvInfoList = append(equipInvInfoList, equipInvInfo)
	}

	return &pbbms.SearchEquipResponse{
		Total: convert.Int64(total),
		List:  equipInvInfoList,
	}, nil
}

// DecreaseEquipInvHandler 扣减设备库存接口
func (h *EquipInventoryHandler) DecreaseEquipInvHandler(ctx context.Context, header *web.Header, i interface{}) (interface{}, *bmserror.BMSError) {
	req := i.(*pbbms.DecreaseEquipInvRequest)
	if req.GetEquipId() == "" {
		return nil, bmserror.NewError(constant.ErrParam, "equip_id is empty")
	}
	if req.GetLabCode() == "" {
		return nil, bmserror.NewError(constant.ErrParam, "lab_code is empty")
	}
	if req.GetCount() <= 0 {
		return nil, bmserror.NewError(constant.ErrParam, "count must be greater than 0")
	}
	if req.GetReason() == "" {
		return nil, bmserror.NewError(constant.ErrParam, "reason is required")
	}

	// 校验设备是否存在
	equipList, _, bmsErr := h.equipMng.SearchEquipMng(ctx, &manager.EquipSearchParam{
		EquipIdList: []string{req.GetEquipId()},
		PageIn:      nil,
	})
	if bmsErr != nil {
		return nil, bmsErr.Mark()
	}
	if len(equipList) == 0 {
		return nil, bmserror.NewError(constant.ErrParam, "equip not found")
	}

	// 获取实验室信息
	lab, bmsErr := h.labMng.GetLabByCode(ctx, req.GetLabCode())
	if bmsErr != nil {
		return nil, bmsErr.Mark()
	}
	if lab == nil {
		return nil, bmserror.NewError(constant.ErrParam, "lab not found")
	}

	// 生成任务ID
	taskId, bmsErr := idutil.GenerateTaskNumber(ctx, constant.EquipInvTaskId)
	if bmsErr != nil {
		return nil, bmsErr.Mark()
	}

	tErr := transaction.PropagationRequired(ctx, func(ctx context.Context) *bmserror.BMSError {
		stockReq := &bmservice.DecreaseStockRequest{
			LabId:          req.GetLabCode(),
			EquipId:        req.GetEquipId(),
			Count:          req.GetCount(),
			SheetId:        taskId,
			TransSheetType: constant.TransactionSheetTypeInventory,
			Operator:       header.UserEmail,
			Description:    req.GetReason(),
		}
		// 调用 service 扣减库存和记录三级账
		_, bmsErr = h.inventoryService.DecreaseInventory(ctx, stockReq)
		if bmsErr != nil {
			return bmsErr.Mark()
		}
		// 创建库存任务
		now := timeutil.GetCurrentUnix()
		task := &entity.InventoryTaskTab{
			TaskID:     taskId,
			TaskType:   constant.InventoryTaskTypeDecrease,
			TaskStatus: constant.InventoryTaskStatusDone,
			LabId:      lab.LabCode,
			EquipId:    req.GetEquipId(),
			TotalQty:   req.GetCount(),
			Operator:   header.UserEmail,
			Remark:     req.GetReason(),
			Ctime:      now,
			Mtime:      now,
		}
		if bmsErr := h.inventoryTaskMng.CreateInventoryTask(ctx, task); bmsErr != nil {
			return bmsErr.Mark()
		}
		// 创建任务日志
		taskLog := &entity.InventoryTaskLogTab{
			TaskID:     taskId,
			TaskStatus: constant.InventoryTaskStatusDone,
			Remark:     req.GetReason(),
			Operator:   header.UserEmail,
		}
		if bmsErr := h.inventoryTaskMng.CreateInventoryTaskLog(ctx, taskLog); bmsErr != nil {
			return bmsErr.Mark()
		}
		return nil
	})
	if tErr != nil {
		return nil, tErr.Mark()
	}
	return &pbbms.DecreaseEquipInvResponse{}, nil
}
