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

func (l LabManager) GetLabByCode(ctx context.Context, labCode string) (*entity.LaboratoryTab, *bmserror.BMSError) {
	var lab entity.LaboratoryTab
	db := l.ds.GetDataSource(ctx, nil).Table(entity.LabTabName).Where("lab_code = ?", labCode).First(&lab)
	if db.RecordNotFound() {
		return nil, nil
	}
	if err := db.GetError(); err != nil {
		return nil, bmserror.NewError(constant.ErrDB, err.Error())
	}
	return &lab, nil
}

// BatchGetLabByCodes 批量获取实验室信息
func (l LabManager) BatchGetLabByCodes(ctx context.Context, labCodes []string) (map[string]*entity.LaboratoryTab, *bmserror.BMSError) {
	if len(labCodes) == 0 {
		return make(map[string]*entity.LaboratoryTab), nil
	}
	var labList []*entity.LaboratoryTab
	if err := l.ds.GetDataSource(ctx, nil).Table(entity.LabTabName).
		Where("lab_code IN (?)", labCodes).
		Find(&labList).GetError(); err != nil {
		return nil, bmserror.NewError(constant.ErrDB, err.Error())
	}
	labMap := make(map[string]*entity.LaboratoryTab)
	for _, lab := range labList {
		labMap[lab.LabCode] = lab
	}
	return labMap, nil
}
