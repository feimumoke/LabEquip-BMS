package manager

import (
	"context"

	"github.com/feimumoke/labequipbms/defines/constant"
	"github.com/feimumoke/labequipbms/defines/entity"
	"github.com/feimumoke/labequipbms/framework/bmserror"
	"github.com/feimumoke/labequipbms/framework/datasource"
)

type LabManager struct {
	ds datasource.DataSource
}

func NewLabManager() *LabManager {
	return &LabManager{ds: datasource.DefaultBasicSource}
}

func (l LabManager) GetAllLabMng(ctx context.Context) ([]*entity.LaboratoryTab, *bmserror.BMSError) {
	var labList []*entity.LaboratoryTab
	db := l.ds.GetDataSource(ctx, nil).Table(entity.LabTabName)
	err := db.Find(&labList).GetError()
	if err != nil {
		return nil, bmserror.NewError(constant.ErrDB, err.Error())
	}
	return labList, nil
}
