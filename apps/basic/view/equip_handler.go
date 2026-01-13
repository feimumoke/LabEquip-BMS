package view

import (
	"context"
	"strings"
	"time"

	pbbasic "github.com/feimumoke/labequipbms/api_idl/apps/basic"
	"github.com/feimumoke/labequipbms/apps/basic/manager"
	"github.com/feimumoke/labequipbms/apps/common/idutil"
	"github.com/feimumoke/labequipbms/defines/constant"
	"github.com/feimumoke/labequipbms/defines/entity"
	"github.com/feimumoke/labequipbms/framework/bmserror"
	"github.com/feimumoke/labequipbms/framework/log"
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
		// 处理图片列表
		var imagesStr string
		if len(req.GetImageList()) > 0 {
			if len(req.GetImageList()) > 10 {
				return bmserror.NewError(constant.ErrParam, "图片数量不能超过10张")
			}
			imagesStr = strings.Join(req.GetImageList(), ",")
		}

		equip := &entity.EquipTab{
			EquipId:      equipNumber,
			CategoryId:   req.GetCategoryId(),
			CategoryName: categoryName,
			EquipName:    req.GetEquipName(),
			Creator:      header.UserEmail,
			Description:  req.GetDescription(),
			Model:        req.GetModel(),
			Images:       imagesStr,
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
		// 解析图片列表

		equipInfo := &pbbasic.EquipInfo{
			EquipId:      &equip.EquipId,
			EquipName:    &equip.EquipName,
			CategoryId:   &equip.CategoryId,
			CategoryName: &equip.CategoryName,
			Description:  &equip.Description,
			Creator:      &equip.Creator,
			Ctime:        &equip.Ctime,
			Model:        &equip.Model,
			Images:       equip.GetImageUrlList(),
		}
		equipInfoList = append(equipInfoList, equipInfo)
	}
	return &pbbasic.SearchEquipResponse{
		Total: convert.Int64(total),
		List:  equipInfoList,
	}, nil
}

func (h EquipHandler) UpdateEquipHandler(ctx context.Context, header *web.Header, i interface{}) (interface{}, *bmserror.BMSError) {
	req := i.(*pbbasic.UpdateEquipRequest)

	if req.GetEquipId() == "" {
		return nil, bmserror.NewError(constant.ErrParam, "equip id is empty")
	}

	// 校验图片数量
	if len(req.GetImageList()) > 10 {
		return nil, bmserror.NewError(constant.ErrParam, "图片数量不能超过10张")
	}

	transactionErr := transaction.PropagationRequired(ctx, func(ctx context.Context) *bmserror.BMSError {
		// 检查设备是否存在
		existEquip, err := h.equipMng.GetEquipById(ctx, req.GetEquipId())
		if err != nil {
			return bmserror.NewError(constant.ErrNotFound, "设备不存在")
		}

		// 构建更新字段
		updates := make(map[string]interface{})
		updates["mtime"] = time.Now().Unix()

		if req.CategoryId != nil {
			updates["category_id"] = req.GetCategoryId()
			// 更新分类名称
			categoryName := constant.GetCategoryName(req.GetCategoryId())
			if categoryName == "" {
				return bmserror.NewError(constant.ErrParam, "category id %v name is empty", req.GetCategoryId())
			}
			updates["category_name"] = categoryName
		}

		if req.EquipName != nil {
			updates["equip_name"] = req.GetEquipName()
		}

		if req.Description != nil {
			updates["description"] = req.GetDescription()
		}

		if req.Model != nil {
			updates["model"] = req.GetModel()
		}

		if len(req.GetImageList()) > 0 {
			// 将图片列表序列化为 JSON 字符串
			updates["images"] = strings.Join(req.GetImageList(), ",")
		}

		// 执行更新
		if err := h.equipMng.UpdateEquip(ctx, req.GetEquipId(), updates); err != nil {
			log.Errorf("update equip error: %v", err)
			return bmserror.NewError(constant.ErrInternalServer, "更新设备失败")
		}

		log.Infof("equip updated successfully: %s, operator: %s, old: %+v, updates: %+v",
			req.GetEquipId(), header.UserEmail, existEquip, updates)

		return nil
	})

	if transactionErr != nil {
		return nil, transactionErr.Mark()
	}

	return &pbbasic.UpdateEquipResponse{}, nil
}
