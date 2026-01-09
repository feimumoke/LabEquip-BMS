package service

import (
	"context"

	"github.com/feimumoke/labequipbms/apps/bms/manager"
	"github.com/feimumoke/labequipbms/apps/common/idutil"
	"github.com/feimumoke/labequipbms/defines/constant"
	"github.com/feimumoke/labequipbms/defines/entity"
	"github.com/feimumoke/labequipbms/framework/bmserror"
	"github.com/feimumoke/labequipbms/framework/support/collection"
	"github.com/feimumoke/labequipbms/framework/support/timeutil"
	"github.com/feimumoke/labequipbms/framework/transaction"
)

type BorrowService struct {
	borrowTaskMng    *manager.BorrowTaskManager
	inventoryService *InventoryService
}

func NewBorrowService() *BorrowService {
	return &BorrowService{
		borrowTaskMng:    manager.NewBorrowTaskManager(),
		inventoryService: NewInventoryService(),
	}
}

// CreateBorrowTask 创建借记任务，尝试预分配库存
// 返回任务ID和任务状态
func (s *BorrowService) CreateBorrowTask(ctx context.Context, labId, equipId string, count int64, description, operator string) (string, constant.BorrowTaskStatus, *bmserror.BMSError) {
	var taskId string
	var taskStatus constant.BorrowTaskStatus

	transactionErr := transaction.PropagationRequired(ctx, func(ctx context.Context) *bmserror.BMSError {
		// 生成任务ID
		var bmsError *bmserror.BMSError
		taskId, bmsError = idutil.GenBorrowTaskId(ctx)
		if bmsError != nil {
			return bmsError.Mark()
		}

		// 检查库存是否足够
		available, bmsError := s.inventoryService.CheckInventoryAvailable(ctx, labId, equipId, count)
		if bmsError != nil {
			return bmsError.Mark()
		}

		now := timeutil.GetCurrentUnix()

		if !available {
			// 库存不足，状态为pending
			taskStatus = constant.BorrowTaskStatusPending
		} else {
			// 预分配库存
			stockReq := &TransStockRequest{
				LabId:          labId,
				EquipId:        equipId,
				Count:          count,
				SheetId:        taskId,
				TransSheetType: constant.TransactionSheetTypeBorrow,
				TransType:      constant.TransactionTypeAllocate,
				Operator:       operator,
				Description:    description,
			}
			_, bmsErr := s.inventoryService.TransInventory(ctx, stockReq)
			if bmsErr != nil {
				return bmsErr.Mark()
			}
			taskStatus = constant.BorrowTaskStatusAllocate
		}
		// 创建借记任务
		task := &entity.BorrowTask{
			TaskID:     taskId,
			EquipId:    equipId,
			LabId:      labId,
			BorrowQty:  count,
			TaskStatus: taskStatus,
			Creator:    operator,
			Ctime:      now,
			Mtime:      now,
		}
		if bmsError := s.borrowTaskMng.CreateBorrowTask(ctx, task); bmsError != nil {
			return bmsError.Mark()
		}

		// 创建任务日志
		statusMsg := "pending"
		if taskStatus == constant.BorrowTaskStatusAllocate {
			statusMsg = "allocated"
		}
		taskLog := &entity.BorrowTaskLog{
			TaskID:     taskId,
			TaskStatus: taskStatus,
			Remark:     description + ", status: " + statusMsg,
			Operator:   operator,
			Ctime:      now,
		}
		if bmsError := s.borrowTaskMng.CreateBorrowTaskLog(ctx, taskLog); bmsError != nil {
			return bmsError.Mark()
		}

		return nil
	})

	if transactionErr != nil {
		return "", 0, transactionErr.Mark()
	}

	return taskId, taskStatus, nil
}

// CancelBorrowTask 取消借记任务，如果已分配库存需要归还
func (s *BorrowService) CancelBorrowTask(ctx context.Context, taskId, reason, operator string) *bmserror.BMSError {
	transactionErr := transaction.PropagationRequired(ctx, func(ctx context.Context) *bmserror.BMSError {
		// 获取借记任务
		task, bmsError := s.borrowTaskMng.GetBorrowTaskByTaskId(ctx, taskId, true)
		if bmsError != nil {
			return bmsError.Mark()
		}
		if task == nil {
			return bmserror.NewError(constant.ErrParam, "borrow task not found")
		}

		// 检查状态，approve之前可以取消
		if task.TaskStatus == constant.BorrowTaskStatusOngoing {
			return bmserror.NewError(constant.ErrParam, "ongoing cannot cancel approved task")
		}
		if collection.Contain(task.TaskStatus, constant.DoneBorrowTaskStatusList) {
			return bmserror.NewError(constant.ErrParam, "final status cannot cancel approved task")
		}

		now := timeutil.GetCurrentUnix()

		// 如果已经分配库存，需要归还库存
		if task.TaskStatus == constant.BorrowTaskStatusAllocate {
			stockReq := &TransStockRequest{
				LabId:          task.LabId,
				EquipId:        task.EquipId,
				Count:          task.BorrowQty,
				SheetId:        taskId,
				TransSheetType: constant.TransactionSheetTypeBorrow,
				TransType:      constant.TransactionTypeReject,
				Operator:       operator,
				Description:    reason,
			}
			_, bmsErr := s.inventoryService.TransInventory(ctx, stockReq)
			if bmsErr != nil {
				return bmsErr.Mark()
			}
		}
		// 更新任务状态
		task.TaskStatus = constant.BorrowTaskStatusCancel
		task.Mtime = now
		if bmsError := s.borrowTaskMng.UpdateBorrowTask(ctx, task); bmsError != nil {
			return bmsError.Mark()
		}

		// 创建任务日志
		if reason == "" {
			reason = "cancelled by user"
		}
		taskLog := &entity.BorrowTaskLog{
			TaskID:     task.TaskID,
			TaskStatus: constant.BorrowTaskStatusCancel,
			Remark:     reason,
			Operator:   operator,
			Ctime:      now,
		}
		if bmsError := s.borrowTaskMng.CreateBorrowTaskLog(ctx, taskLog); bmsError != nil {
			return bmsError.Mark()
		}

		return nil
	})

	if transactionErr != nil {
		return transactionErr.Mark()
	}

	return nil
}

// ApproveBorrowTask 审批借记任务，库存转为借记库存
func (s *BorrowService) ApproveBorrowTask(ctx context.Context, taskId string, approved bool, reason, operator string) *bmserror.BMSError {
	transactionErr := transaction.PropagationRequired(ctx, func(ctx context.Context) *bmserror.BMSError {
		// 获取借记任务
		task, bmsError := s.borrowTaskMng.GetBorrowTaskByTaskId(ctx, taskId, true)
		if bmsError != nil {
			return bmsError.Mark()
		}
		if task == nil {
			return bmserror.NewError(constant.ErrParam, "borrow task not found")
		}
		// 检查状态，必须是已经分配库存状态
		if task.TaskStatus != constant.BorrowTaskStatusAllocate {
			return bmserror.NewError(constant.ErrParam, "task status must be allocated")
		}
		task.Approval = operator

		now := timeutil.GetCurrentUnix()
		if approved {
			// 审批通过，通知用户
			// 更新任务状态
			task.TaskStatus = constant.BorrowTaskStatusApproval
		} else {
			// 审批拒绝，归还预分配库存
			stockReq := &TransStockRequest{
				LabId:          task.LabId,
				EquipId:        task.EquipId,
				Count:          task.BorrowQty,
				SheetId:        taskId,
				TransSheetType: constant.TransactionSheetTypeBorrow,
				TransType:      constant.TransactionTypeReject,
				Operator:       operator,
				Description:    reason,
			}
			_, bmsErr := s.inventoryService.TransInventory(ctx, stockReq)
			if bmsErr != nil {
				return bmsErr.Mark()
			}
			// 更新任务状态
			task.TaskStatus = constant.BorrowTaskStatusReject
		}

		task.Mtime = now
		if bmsError := s.borrowTaskMng.UpdateBorrowTask(ctx, task); bmsError != nil {
			return bmsError.Mark()
		}
		// 创建任务日志
		if reason == "" {
			if approved {
				reason = "approved"
			} else {
				reason = "rejected"
			}
		}
		taskLog := &entity.BorrowTaskLog{
			TaskID:     task.TaskID,
			TaskStatus: task.TaskStatus,
			Remark:     reason,
			Operator:   operator,
			Ctime:      now,
		}
		if bmsError := s.borrowTaskMng.CreateBorrowTaskLog(ctx, taskLog); bmsError != nil {
			return bmsError.Mark()
		}

		return nil
	})

	if transactionErr != nil {
		return transactionErr.Mark()
	}

	return nil
}

// TaskBorrowTask 拿走借记设备
func (s *BorrowService) TaskBorrowTask(ctx context.Context, taskId, operator string, codeList []string) *bmserror.BMSError {
	transactionErr := transaction.PropagationRequired(ctx, func(ctx context.Context) *bmserror.BMSError {
		// 查找相关的借记任务并完结
		// 查找状态为已审批的任务
		task, bmsError := s.borrowTaskMng.GetBorrowTaskByTaskId(ctx, taskId, true)
		if bmsError != nil {
			return bmsError.Mark()
		}
		if task == nil {
			return bmserror.NewError(constant.ErrParam, "borrow task not found")
		}
		if task.TaskStatus != constant.BorrowTaskStatusApproval {
			return bmserror.NewError(constant.ErrParam, "task status must be approved")
		}
		//归还借记库存
		stockReq := &TransStockRequest{
			LabId:          task.LabId,
			EquipId:        task.EquipId,
			Count:          task.BorrowQty,
			SheetId:        taskId,
			TransSheetType: constant.TransactionSheetTypeBorrow,
			TransType:      constant.TransactionTypeBorrow,
			Operator:       operator,
			Description:    "task borrow",
		}
		_, bmsErr := s.inventoryService.TransInventory(ctx, stockReq)
		if bmsErr != nil {
			return bmsErr.Mark()
		}

		task.TaskStatus = constant.BorrowTaskStatusOngoing
		task.Mtime = timeutil.GetCurrentUnix()
		if bmsError := s.borrowTaskMng.UpdateBorrowTask(ctx, task); bmsError != nil {
			return bmsError.Mark()
		}
		// 创建任务日志
		taskLog := &entity.BorrowTaskLog{
			TaskID:     task.TaskID,
			TaskStatus: constant.BorrowTaskStatusOngoing,
			Remark:     "returned",
			Operator:   operator,
			Ctime:      timeutil.GetCurrentUnix(),
		}
		if bmsError := s.borrowTaskMng.CreateBorrowTaskLog(ctx, taskLog); bmsError != nil {
			return bmsError.Mark()
		}
		return nil
	})

	if transactionErr != nil {
		return transactionErr.Mark()
	}

	return nil
}

// CompleteBorrowTask 完成借记任务（归还后调用）
func (s *BorrowService) CompleteBorrowTask(ctx context.Context, taskId, operator string, returnQty int64) *bmserror.BMSError {
	transactionErr := transaction.PropagationRequired(ctx, func(ctx context.Context) *bmserror.BMSError {
		// 查找相关的借记任务并完结
		// 查找状态为已审批的任务
		task, bmsError := s.borrowTaskMng.GetBorrowTaskByTaskId(ctx, taskId, true)
		if bmsError != nil {
			return bmsError.Mark()
		}
		if task == nil {
			return bmserror.NewError(constant.ErrParam, "borrow task not found")
		}
		if task.IsDone() {
			return bmserror.NewError(constant.ErrParam, "task is done")
		}
		if task.BorrowQty != returnQty {
			return bmserror.NewError(constant.ErrParam, "return qty is invalid")
		}
		if task.TaskStatus != constant.BorrowTaskStatusOngoing {
			return bmserror.NewError(constant.ErrParam, "task status must be ongoing")
		}
		//归还借记库存
		stockReq := &TransStockRequest{
			LabId:          task.LabId,
			EquipId:        task.EquipId,
			Count:          task.BorrowQty,
			SheetId:        taskId,
			TransSheetType: constant.TransactionSheetTypeBorrow,
			TransType:      constant.TransactionTypeReturn,
			Operator:       operator,
			Description:    "return",
		}
		_, bmsErr := s.inventoryService.TransInventory(ctx, stockReq)
		if bmsErr != nil {
			return bmsErr.Mark()
		}

		task.TaskStatus = constant.BorrowTaskStatusDone
		task.Mtime = timeutil.GetCurrentUnix()
		if bmsError := s.borrowTaskMng.UpdateBorrowTask(ctx, task); bmsError != nil {
			return bmsError.Mark()
		}
		// 创建任务日志
		taskLog := &entity.BorrowTaskLog{
			TaskID:     task.TaskID,
			TaskStatus: constant.BorrowTaskStatusDone,
			Remark:     "returned",
			Operator:   operator,
			Ctime:      timeutil.GetCurrentUnix(),
		}
		if bmsError := s.borrowTaskMng.CreateBorrowTaskLog(ctx, taskLog); bmsError != nil {
			return bmsError.Mark()
		}
		return nil
	})

	if transactionErr != nil {
		return transactionErr.Mark()
	}

	return nil
}
