package view

import (
	"context"

	pbbasic "github.com/feimumoke/labequipbms/api_idl/apps/basic"
	"github.com/feimumoke/labequipbms/apps/basic/manager"
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

type EquipHandler struct {
	equipMng *manager.EquipManager
}

func NewEquipHandler() *EquipHandler {
	return &EquipHandler{equipMng: manager.NewEquipManager()}
}

func (h EquipHandler) CreateEquipHandler(ctx context.Context, header *web.Header, i interface{}) (interface{}, *bmserror.BMSError) {
	req := i.(*pbbasic.CreateEquipRequest)
	if req.GetCategoryId() == 0 {
		return nil, bmserror.NewError(constant.ErrParam, "category id is empty")
	}
	if req.GetEquipName() == "" {
		return nil, bmserror.NewError(constant.ErrParam, "equip name id is empty")
	}
	categoryName := constant.GetCategoryName(req.GetCategoryId())
	if categoryName == "" {
		return nil, bmserror.NewError(constant.ErrParam, "category id %v name is empty", req.GetCategoryId())
	}
	now := timeutil.GetCurrentUnix()
	transactionErr := transaction.PropagationRequired(ctx, func(ctx context.Context) *bmserror.BMSError {
		equipList, _, bmsError := h.equipMng.SearchEquipMng(ctx, &manager.EquipSearchParam{
			EquipName: req.GetEquipName(),
			WithLock:  true,
			PageIn:    nil,
		})
		if bmsError != nil {
			return bmsError.Mark()
		}
		if len(equipList) > 0 {
			return bmserror.NewError(constant.ErrParam, "equip %v already exists", req.GetEquipName())
		}
		equipNumber, bmsError := idutil.GenEquipNumber(ctx)
		if bmsError != nil {
			return bmsError.Mark()
		}
		equip := &entity.EquipTab{
			EquipId:      equipNumber,
			CategoryId:   req.GetCategoryId(),
			CategoryName: categoryName,
			EquipName:    req.GetEquipName(),
			Creator:      header.UserEmail,
			Description:  req.GetDescription(),
			Ctime:        now,
			Mtime:        now,
		}

		cErr := h.equipMng.CreateEquip(ctx, equip)
		if cErr != nil {
			return cErr.Mark()
		}
		return nil
	})
	if transactionErr != nil {
		return nil, transactionErr.Mark()
	}
	return &pbbasic.CreateEquipResponse{}, nil
}

func (h EquipHandler) SearchEquipHandler(ctx context.Context, header *web.Header, i interface{}) (interface{}, *bmserror.BMSError) {
	req := i.(*pbbasic.SearchEquipRequest)
	pageIn := &paginator.PageIn{
		Pageno:     req.GetPageno(),
		Count:      req.GetCount(),
		IsGetTotal: true,
	}
	equipList, total, bmsError := h.equipMng.SearchEquipMng(ctx, &manager.EquipSearchParam{
		CategoryIdList: req.GetCategoryIdList(),
		EquipName:      req.GetEquipName(),
		PageIn:         pageIn,
	})
	if bmsError != nil {
		return nil, bmsError.Mark()
	}
	equipInfoList := make([]*pbbasic.EquipInfo, 0, len(equipList))
	for _, equip := range equipList {
		equipInfo := &pbbasic.EquipInfo{
			EquipId:      &equip.EquipId,
			EquipName:    &equip.EquipName,
			CategoryId:   &equip.CategoryId,
			CategoryName: &equip.CategoryName,
			Description:  &equip.Description,
			Creator:      &equip.Creator,
			Ctime:        &equip.Ctime,
		}
		equipInfoList = append(equipInfoList, equipInfo)
	}
	return &pbbasic.SearchEquipResponse{
		Total: convert.Int64(total),
		List:  equipInfoList,
	}, nil
}
