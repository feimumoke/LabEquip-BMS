package view

import (
	"context"

	pbbms "github.com/feimumoke/labequipbms/api_idl/apps/bms"
	"github.com/feimumoke/labequipbms/apps/basic/manager"
	bmservice "github.com/feimumoke/labequipbms/apps/bms/service"
	"github.com/feimumoke/labequipbms/defines/constant"
	"github.com/feimumoke/labequipbms/defines/message"
	"github.com/feimumoke/labequipbms/framework/bmserror"
	"github.com/feimumoke/labequipbms/framework/web"
)

type ApprovalHandler struct {
	borrowService *bmservice.BorrowService
	userMng       *manager.UserManager
}

func NewApprovalHandler() *ApprovalHandler {
	return &ApprovalHandler{
		borrowService: bmservice.NewBorrowService(),
		userMng:       manager.NewUserManager(),
	}
}

// ApproveBorrowHandler 借记Approve接口
func (h *ApprovalHandler) ApproveBorrowHandler(ctx context.Context, header *web.Header, i interface{}) (interface{}, *bmserror.BMSError) {
	req := i.(*pbbms.ApproveBorrowRequest)
	if req.GetBorrowId() == "" {
		return nil, bmserror.NewError(constant.ErrParam, "borrow_id is empty")
	}
	if req.Approved == nil {
		return nil, bmserror.NewError(constant.ErrParam, "approved is required")
	}
	approved := req.GetApproved() == 1

	// 检查用户是否是老师
	user, bmsErr := h.userMng.GetUserByEmail(ctx, header.UserEmail)
	if bmsErr != nil {
		return nil, bmsErr.Mark()
	}
	if user == nil {
		return nil, bmserror.NewError(constant.ErrParam, "user not found")
	}
	if user.Role != constant.UserRoleTypeTeacher && user.Role != constant.UserRoleTypeAdmin && user.Role != constant.UserRoleTypeSuperAdmin {
		return nil, bmserror.NewError(constant.ErrParam, "only teacher can approve")
	}

	// 调用 service 审批借记任务
	bmsErr = h.borrowService.ApproveBorrowTask(ctx, req.GetBorrowId(), approved, req.GetReason(), header.UserEmail)
	if bmsErr != nil {
		return nil, bmsErr.Mark()
	}

	return &pbbms.ApproveBorrowResponse{}, nil
}

func (h *ApprovalHandler) ApprovalMessage(ctx context.Context, msg interface{}) *bmserror.BMSError {
	req := msg.(*message.ApproveBorrowMessage)
	return h.borrowService.ApprovalMessage(ctx, req)
}
