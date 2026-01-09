package manager

import (
	"context"

	"github.com/feimumoke/labequipbms/defines/constant"
	"github.com/feimumoke/labequipbms/defines/entity"
	"github.com/feimumoke/labequipbms/framework/bmserror"
	"github.com/feimumoke/labequipbms/framework/datasource"
	"github.com/feimumoke/labequipbms/framework/support/paginator"
)

type TransactionManager struct {
	ds datasource.DataSource
}

func NewTransactionManager() *TransactionManager {
	return &TransactionManager{ds: datasource.DefaultInvSource}
}

// CreateTransactionLog 创建三级账流水
func (m *TransactionManager) CreateTransactionLog(ctx context.Context, log *entity.TransactionLogTab) *bmserror.BMSError {
	if err := m.ds.GetDataSource(ctx, nil).Table(entity.TransactionLogTabName).Create(log).GetError(); err != nil {
		return bmserror.NewError(constant.ErrDB, err.Error())
	}
	return nil
}

// SearchTransactionLog 查询三级账流水
type TransactionSearchParam struct {
	LabCode  string
	EquipId  string
	Operator string
	PageIn   *paginator.PageIn
}

func (m *TransactionManager) SearchTransactionLog(ctx context.Context, params *TransactionSearchParam) ([]*entity.TransactionLogTab, int64, *bmserror.BMSError) {
	var logList []*entity.TransactionLogTab
	db := m.ds.GetDataSource(ctx, nil).Table(entity.TransactionLogTabName)
	if params.LabCode != "" {
		db = db.Where("lab_id = ?", params.LabCode)
	}
	if params.EquipId != "" {
		db = db.Where("equip_id = ?", params.EquipId)
	}
	if params.Operator != "" {
		db = db.Where("operator = ?", params.Operator)
	}
	db = db.Order("ctime DESC")
	total, err := paginator.Paginator(db, params.PageIn, &logList)
	if err != nil {
		return nil, 0, err.Mark()
	}
	return logList, total, nil
}
