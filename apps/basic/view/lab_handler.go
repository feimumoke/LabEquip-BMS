package view

import (
	"context"

	pbbasic "github.com/feimumoke/labequipbms/api_idl/apps/basic"
	"github.com/feimumoke/labequipbms/apps/basic/manager"
	"github.com/feimumoke/labequipbms/framework/bmserror"
	"github.com/feimumoke/labequipbms/framework/web"
)

type LabHandler struct {
	labMng *manager.LabManager
}

func NewLabHandler() *LabHandler {
	return &LabHandler{labMng: manager.NewLabManager()}
}

func (h LabHandler) SearchLabHandler(ctx context.Context, header *web.Header, i interface{}) (interface{}, *bmserror.BMSError) {
	labList, bmsError := h.labMng.GetAllLabMng(ctx)
	if bmsError != nil {
		return nil, bmsError.Mark()
	}
	labInfoList := make([]*pbbasic.LabInfo, 0, len(labList))
	for _, lab := range labList {
		labInfo := &pbbasic.LabInfo{
			LabCode:     &lab.LabCode,
			LabName:     &lab.LabName,
			Address:     &lab.Address,
			Description: &lab.Description,
			Creator:     nil, // LaboratoryTab 中没有 Creator 字段
			Ctime:       &lab.Ctime,
		}
		labInfoList = append(labInfoList, labInfo)
	}
	total := int64(len(labInfoList))
	return &pbbasic.SearchLabResponse{
		Total: &total,
		List:  labInfoList,
	}, nil
}
