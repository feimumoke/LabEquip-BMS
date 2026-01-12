package view

import (
	pbbms "github.com/feimumoke/labequipbms/api_idl/apps/bms"
	"github.com/feimumoke/labequipbms/defines/message"
	"github.com/feimumoke/labequipbms/framework/asynctask"
	"github.com/feimumoke/labequipbms/framework/web"
)

func InitBMSView(s *web.BasicServer, r *asynctask.AsyncRunner) {
	initEquipInventory(s, r)
	initBorrowing(s, r)

	initApproval(s, r)
	initInventoryTask(s, r)
	initTransaction(s, r)
}

func initEquipInventory(s *web.BasicServer, r *asynctask.AsyncRunner) {
	h := NewEquipInventoryHandler()
	s.RegisterPOST("/apps/bms/inventory/create_equip_inv", h.CreateEquipInvHandler, &pbbms.CreateEquipInvRequest{})
	s.RegisterPOST("/apps/bms/inventory/search_equip_inv", h.SearchEquipInvHandler, &pbbms.SearchEquipRequest{})
	s.RegisterPOST("/apps/bms/inventory/decrease_equip_inv", h.DecreaseEquipInvHandler, &pbbms.DecreaseEquipInvRequest{})
}

func initInventoryTask(s *web.BasicServer, r *asynctask.AsyncRunner) {
	h := NewInventoryTaskHandler()
	s.RegisterPOST("/apps/bms/inventory/search_inv_task", h.SearchInvTaskHandler, &pbbms.SearchInvTaskRequest{})
}

func initBorrowing(s *web.BasicServer, r *asynctask.AsyncRunner) {
	h := NewBorrowingHandler()
	s.RegisterPOST("/apps/bms/borrow/create_borrow", h.CreateBorrowHandler, &pbbms.CreateBorrowRequest{})
	s.RegisterPOST("/apps/bms/borrow/cancel_borrow", h.CancelBorrowHandler, &pbbms.CancelBorrowRequest{})
	s.RegisterPOST("/apps/bms/borrow/task_borrow", h.TaskBorrowHandler, &pbbms.TaskBorrowRequest{})
	s.RegisterPOST("/apps/bms/borrow/return_borrow", h.ReturnBorrowHandler, &pbbms.ReturnBorrowRequest{})
	s.RegisterPOST("/apps/bms/borrow/search_borrow_task", h.SearchBorrowTaskHandler, &pbbms.SearchBorrowTaskRequest{})
}

func initApproval(s *web.BasicServer, r *asynctask.AsyncRunner) {
	h := NewApprovalHandler()
	r.RegisterMessageHandler(message.ApproveBorrowTaskName, h.ApprovalMessage, &message.ApproveBorrowMessage{})
	s.RegisterPOST("/apps/bms/borrow/approve_borrow", h.ApproveBorrowHandler, &pbbms.ApproveBorrowRequest{})
}

func initTransaction(s *web.BasicServer, r *asynctask.AsyncRunner) {
	h := NewTransactionHandler()
	s.RegisterPOST("/apps/bms/transaction/search_inv_transaction", h.SearchInvTransactionHandler, &pbbms.SearchInvTransactionRequest{})
}
