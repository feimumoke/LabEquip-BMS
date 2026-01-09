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

type TransactionHandler struct {
	transactionMng *bmsmanager.TransactionManager
	equipMng       *manager.EquipManager
	labMng         *manager.LabManager
}

func NewTransactionHandler() *TransactionHandler {
	return &TransactionHandler{
		transactionMng: bmsmanager.NewTransactionManager(),
		equipMng:       manager.NewEquipManager(),
		labMng:         manager.NewLabManager(),
	}
}

// SearchInvTransactionHandler 三级账查询接口
func (h *TransactionHandler) SearchInvTransactionHandler(ctx context.Context, header *web.Header, i interface{}) (interface{}, *bmserror.BMSError) {
	req := i.(*pbbms.SearchInvTransactionRequest)
	pageIn := &paginator.PageIn{
		Pageno:     req.GetPageno(),
		Count:      req.GetCount(),
		IsGetTotal: true,
	}

	transList, total, bmsErr := h.transactionMng.SearchTransactionLog(ctx, &bmsmanager.TransactionSearchParam{
		LabCode:  req.GetLabCode(),
		EquipId:  req.GetEquipId(),
		Operator: req.GetOperator(),
		PageIn:   pageIn,
	})
	if bmsErr != nil {
		return nil, bmsErr.Mark()
	}

	// 获取设备信息
	equipMap := make(map[string]*entity.EquipTab)
	if len(transList) > 0 {
		equipIdList := make([]string, 0, len(transList))
		for _, trans := range transList {
			equipIdList = append(equipIdList, trans.EquipId)
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
	for _, trans := range transList {
		labCodeSet[trans.LabId] = true
	}
	labCodeList := make([]string, 0, len(labCodeSet))
	for labCode := range labCodeSet {
		labCodeList = append(labCodeList, labCode)
	}
	labMap, bmsErr := h.labMng.BatchGetLabByCodes(ctx, labCodeList)
	if bmsErr != nil {
		return nil, bmsErr.Mark()
	}

	invTransactionList := make([]*pbbms.InvTransaction, 0, len(transList))
	for _, trans := range transList {
		equip := equipMap[trans.EquipId]
		lab := labMap[trans.LabId]

		equipName := ""
		if equip != nil {
			equipName = equip.EquipName
		}
		labName := ""
		if lab != nil {
			labName = lab.LabName
		}

		invTransaction := &pbbms.InvTransaction{
			TransactionId: &trans.TransactionID,
			SheetId:       &trans.SheetID,
			EquipId:       &trans.EquipId,
			EquipName:     []string{equipName},
			LabCode:       &trans.LabId,
			LabName:       &labName,
			Operator:      &trans.Operator,
			Remark:        &trans.Remark,
			AvailQty:      &trans.AvailQty,
			OpQty:         &trans.OpQty,
			TransType:     convert.Int64(trans.TransType),
			SheetType:     convert.Int64(trans.SheetType),
			Ctime:         &trans.Ctime,
		}
		invTransactionList = append(invTransactionList, invTransaction)
	}

	return &pbbms.SearchInvTransactionResponse{
		Total: convert.Int64(total),
		List:  invTransactionList,
	}, nil
}
