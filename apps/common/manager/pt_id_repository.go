package cmanager

import (
	"context"
	"time"

	"github.com/feimumoke/labequipbms/defines/constant"
	"github.com/feimumoke/labequipbms/framework/bmserror"
	"github.com/feimumoke/labequipbms/framework/orm"
)

func (i *Generator) getOrm(ctx context.Context) orm.GORM {
	return i.ds.GetDataSource(ctx, nil)
}

func (i *Generator) create(ctx context.Context, id *DistributedIDcreatorTab) *bmserror.BMSError {
	err := i.getOrm(ctx).Table(DistributedIDcreatorTabTableName).Create(id).GetError()
	if err != nil {
		return err
	}
	return nil
}
func (i *Generator) search(ctx context.Context, idType int64, datePeriod int64, ptNo PtNoType) ([]*DistributedIDcreatorTab, *bmserror.BMSError) {
	var idList []*DistributedIDcreatorTab
	whereCondition := "id_type = ? AND date_period = ? AND pt_no = ?"
	err := i.getOrm(ctx).Table(DistributedIDcreatorTabTableName).Where(whereCondition, idType, datePeriod, ptNo).Find(&idList).GetError()
	if err != nil {
		return nil, bmserror.NewError(constant.ErrDB, err.Error())
	}
	return idList, nil
}
func (i *Generator) update(ctx context.Context, idType int64, datePeriod int64, ptNo PtNoType, originalOffset int64, newOffset int64) (int64, *bmserror.BMSError) {
	whereCondition := "id_value=? AND id_type = ? AND date_period = ? AND pt_no = ?"
	ret := i.getOrm(ctx).Table(DistributedIDcreatorTabTableName).Where(whereCondition, originalOffset, idType, datePeriod,
		ptNo).Updates(&DistributedIDcreatorTab{
		IDValue: newOffset,
		Mtime:   time.Now().Unix(),
	})
	err := ret.GetError()
	if err != nil {
		return 0, bmserror.NewError(constant.ErrDB, err.Error())
	}
	return ret.RowsAffected(), nil
}

func (i *Generator) delete(ctx context.Context, idType int64, mtime int64) *bmserror.BMSError {
	whereCondition := "id_value=? AND mtime <= ?"
	err := i.getOrm(ctx).Table(DistributedIDcreatorTabTableName).Where(whereCondition, idType, mtime).Delete(&DistributedIDcreatorTab{}).GetError()
	if err != nil {
		return bmserror.NewError(constant.ErrDB, err.Error())
	}
	return nil
}
