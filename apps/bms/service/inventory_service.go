package service

import (
	"context"

	"github.com/feimumoke/labequipbms/apps/bms/manager"
	"github.com/feimumoke/labequipbms/apps/common/idutil"
	"github.com/feimumoke/labequipbms/defines/constant"
	"github.com/feimumoke/labequipbms/defines/entity"
	"github.com/feimumoke/labequipbms/framework/bmserror"
	"github.com/feimumoke/labequipbms/framework/support/timeutil"
	"github.com/feimumoke/labequipbms/framework/transaction"
)

type InventoryService struct {
	inventoryMng   *manager.InventoryManager
	transactionMng *manager.TransactionManager
}

func NewInventoryService() *InventoryService {
	return &InventoryService{
		inventoryMng:   manager.NewInventoryManager(),
		transactionMng: manager.NewTransactionManager(),
	}
}

type IncreaseStockRequest struct {
	LabId          string
	EquipId        string
	Count          int64
	SheetId        string
	TransSheetType constant.TransactionSheetType
	Operator       string
	Description    string
}

type IncreaseStockResponse struct {
	TransSheetID string `json:"trans_sheet_id"`
}

// IncreaseInventory 增加库存，同时记录三级账流水
func (s *InventoryService) IncreaseInventory(ctx context.Context, req *IncreaseStockRequest) (*IncreaseStockResponse, *bmserror.BMSError) {
	labId, equipId, count, operator := req.LabId, req.EquipId, req.Count, req.Operator
	if labId == "" || equipId == "" {
		return nil, bmserror.NewError(constant.ErrParam, "lab code and equip code can not be empty")
	}
	var inventory *entity.InventoryTab
	var transactionId string

	transactionErr := transaction.PropagationRequired(ctx, func(ctx context.Context) *bmserror.BMSError {
		// 创建或更新库存
		var bmsError *bmserror.BMSError
		inventory, bmsError = s.inventoryMng.CreateOrUpdateInventory(ctx, labId, equipId, count, operator)
		if bmsError != nil {
			return bmsError.Mark()
		}
		// 生成交易ID
		transactionId, bmsError = idutil.GenerateTransactionID(ctx)
		if bmsError != nil {
			return bmsError.Mark()
		}
		// 创建三级账流水
		transLog := &entity.TransactionLogTab{
			TransactionID: transactionId,
			SheetID:       req.SheetId,
			EquipId:       equipId,
			LabId:         labId,
			Operator:      operator,
			Remark:        req.Description,
			TotalQty:      inventory.TotalQty,
			OnHandQty:     inventory.OnHandQty,
			AvailableQty:  inventory.AvailableQty,
			BorrowedQty:   inventory.BorrowedQty,
			AllocatedQty:  inventory.AllocatedQty,
			OpQty:         count,
			TransType:     constant.TransactionTypeIncrease,
			SheetType:     req.TransSheetType,
			Ctime:         timeutil.GetCurrentUnix(),
		}
		if bmsError := s.transactionMng.CreateTransactionLog(ctx, transLog); bmsError != nil {
			return bmsError.Mark()
		}

		return nil
	})

	if transactionErr != nil {
		return nil, transactionErr.Mark()
	}
	return &IncreaseStockResponse{TransSheetID: transactionId}, nil
}

type DecreaseStockRequest struct {
	LabId          string
	EquipId        string
	Count          int64
	SheetId        string
	TransSheetType constant.TransactionSheetType
	Operator       string
	Description    string
}

type DecreaseStockResponse struct {
	TransSheetID string `json:"trans_sheet_id"`
}

// DecreaseInventory 扣减库存，同时记录三级账流水
func (s *InventoryService) DecreaseInventory(ctx context.Context, req *DecreaseStockRequest) (*DecreaseStockResponse, *bmserror.BMSError) {
	labId, equipId, count, operator := req.LabId, req.EquipId, req.Count, req.Operator

	var inventory *entity.InventoryTab
	var transactionId string
	var bmsError *bmserror.BMSError

	if labId == "" || equipId == "" {
		return nil, bmserror.NewError(constant.ErrParam, "lab code and equip code can not be empty")
	}

	transactionErr := transaction.PropagationRequired(ctx, func(ctx context.Context) *bmserror.BMSError {
		// 扣减库存
		if bmsError := s.inventoryMng.DecreaseInventory(ctx, labId, equipId, count, operator); bmsError != nil {
			return bmsError.Mark()
		}
		// 获取更新后的库存
		inventory, bmsError = s.inventoryMng.GetInventoryByLabAndEquip(ctx, labId, equipId, false)
		if bmsError != nil {
			return bmsError.Mark()
		}

		// 生成交易ID
		transactionId, bmsError = idutil.GenerateTransactionID(ctx)
		if bmsError != nil {
			return bmsError.Mark()
		}
		// 创建三级账流水
		transLog := &entity.TransactionLogTab{
			TransactionID: transactionId,
			SheetID:       req.SheetId,
			EquipId:       equipId,
			LabId:         labId,
			Operator:      operator,
			Remark:        req.Description,
			TotalQty:      inventory.TotalQty,
			OnHandQty:     inventory.OnHandQty,
			AvailableQty:  inventory.AvailableQty,
			BorrowedQty:   inventory.BorrowedQty,
			AllocatedQty:  inventory.AllocatedQty,
			OpQty:         count,
			TransType:     constant.TransactionTypeDecrease,
			SheetType:     req.TransSheetType,
			Ctime:         timeutil.GetCurrentUnix(),
		}
		if bmsError := s.transactionMng.CreateTransactionLog(ctx, transLog); bmsError != nil {
			return bmsError.Mark()
		}

		return nil
	})

	if transactionErr != nil {
		return nil, transactionErr.Mark()
	}

	return &DecreaseStockResponse{TransSheetID: transactionId}, nil
}

type TransStockRequest struct {
	LabId          string
	EquipId        string
	Count          int64
	SheetId        string
	TransSheetType constant.TransactionSheetType
	TransType      constant.TransactionType
	Operator       string
	Description    string
}

type TransStockResponse struct {
	TransSheetID string `json:"trans_sheet_id"`
}

// TransInventory 库存分配 借用 归还
func (s *InventoryService) TransInventory(ctx context.Context, req *TransStockRequest) (*TransStockResponse, *bmserror.BMSError) {
	labId, equipId, count, operator := req.LabId, req.EquipId, req.Count, req.Operator

	var inventory *entity.InventoryTab
	var transactionId string
	var bmsError *bmserror.BMSError

	if labId == "" || equipId == "" {
		return nil, bmserror.NewError(constant.ErrParam, "lab code and equip code can not be empty")
	}

	transactionErr := transaction.PropagationRequired(ctx, func(ctx context.Context) *bmserror.BMSError {
		// 获取库存
		inventory, bmsError = s.inventoryMng.GetInventoryByLabAndEquip(ctx, labId, equipId, true)
		if bmsError != nil {
			return bmsError.Mark()
		}
		if inventory == nil {
			return bmserror.NewError(constant.ErrParam, "inventory not found")
		}

		switch req.TransType {
		case constant.TransactionTypeAllocate: //分配
			if inventory.AvailableQty < count {
				return bmserror.NewError(constant.ErrParam, "insufficient available qty %v total %v", count, inventory.AvailableQty)
			}
			inventory.AvailableQty -= count
			inventory.AllocatedQty += count
		case constant.TransactionTypeBorrow:
			if inventory.OnHandQty < count {
				return bmserror.NewError(constant.ErrParam, "insufficient on hand qty %v total %v", count, inventory.OnHandQty)
			}
			if inventory.AllocatedQty < count {
				return bmserror.NewError(constant.ErrParam, "insufficient allocated qty %v total %v", count, inventory.OnHandQty)
			}
			inventory.OnHandQty -= count
			inventory.AllocatedQty -= count
			inventory.BorrowedQty += count
		case constant.TransactionTypeReturn:
			if inventory.BorrowedQty < count {
				return bmserror.NewError(constant.ErrParam, "insufficient borrow qty %v total %v", count, inventory.OnHandQty)
			}
			inventory.AvailableQty += count
			inventory.OnHandQty += count
			inventory.BorrowedQty -= count
		case constant.TransactionTypeReject: //拒绝 分配后
			if inventory.AllocatedQty < count {
				return bmserror.NewError(constant.ErrParam, "insufficient borrow qty %v total %v", count, inventory.OnHandQty)
			}
			inventory.AvailableQty += count
			inventory.AllocatedQty -= count
		default:
			return bmserror.NewError(constant.ErrParam, "invalid transaction type")
		}

		// 归还库存：从BorrowedQty转为AvailableQty
		if inventory.BorrowedQty < count {
			return bmserror.NewError(constant.ErrParam, "borrowed quantity is not enough")
		}
		inventory.BorrowedQty -= count
		inventory.AvailableQty += count
		inventory.Operator = operator
		inventory.Mtime = timeutil.GetCurrentUnix()
		if bmsError := s.inventoryMng.UpdateInventory(ctx, inventory); bmsError != nil {
			return bmsError.Mark()
		}

		// 生成交易ID
		transactionId, bmsError = idutil.GenerateTransactionID(ctx)
		if bmsError != nil {
			return bmsError.Mark()
		}

		// 创建三级账流水
		transLog := &entity.TransactionLogTab{
			TransactionID: transactionId,
			SheetID:       req.SheetId,
			EquipId:       equipId,
			LabId:         labId,
			Operator:      operator,
			Remark:        req.Description,
			TotalQty:      inventory.TotalQty,
			OnHandQty:     inventory.OnHandQty,
			AvailableQty:  inventory.AvailableQty,
			BorrowedQty:   inventory.BorrowedQty,
			AllocatedQty:  inventory.AllocatedQty,
			OpQty:         count,
			TransType:     constant.TransactionTypeIncrease,
			SheetType:     constant.TransactionSheetTypeBorrow,
			Ctime:         timeutil.GetCurrentUnix(),
		}
		if bmsError := s.transactionMng.CreateTransactionLog(ctx, transLog); bmsError != nil {
			return bmsError.Mark()
		}
		return nil
	})
	if transactionErr != nil {
		return nil, transactionErr.Mark()
	}
	return &TransStockResponse{TransSheetID: transactionId}, nil
}

// CheckInventoryAvailable 检查库存是否足够（不锁定，只读查询）
func (s *InventoryService) CheckInventoryAvailable(ctx context.Context, labId, equipId string, count int64) (bool, *bmserror.BMSError) {
	inventory, bmsError := s.inventoryMng.GetInventoryByLabAndEquip(ctx, labId, equipId, false)
	if bmsError != nil {
		return false, bmsError.Mark()
	}
	if inventory == nil {
		return false, nil
	}
	return inventory.AvailableQty >= count, nil
}
