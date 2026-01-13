package manager

import (
	"context"

	"github.com/feimumoke/labequipbms/defines/entity"
	"github.com/feimumoke/labequipbms/framework/bmserror"
	"github.com/feimumoke/labequipbms/framework/datasource"
	"github.com/feimumoke/labequipbms/framework/support/paginator"
)

type EquipManager struct {
	ds datasource.DataSource
}

func NewEquipManager() *EquipManager {
	return &EquipManager{ds: datasource.DefaultBasicSource}
}

func (m *EquipManager) CreateEquip(ctx context.Context, equip *entity.EquipTab) *bmserror.BMSError {
	if err := m.ds.GetDataSource(ctx, nil).Table(entity.EquipTabName).Create(equip).GetError(); err != nil {
		return err.Mark()
	}
	return nil
}

func (m *EquipManager) UpdateEquip(ctx context.Context, equipId string, updates map[string]interface{}) *bmserror.BMSError {
	if err := m.ds.GetDataSource(ctx, nil).Table(entity.EquipTabName).
		Where("equip_id = ?", equipId).
		Updates(updates).GetError(); err != nil {
		return err.Mark()
	}
	return nil
}

func (m *EquipManager) GetEquipById(ctx context.Context, equipId string) (*entity.EquipTab, *bmserror.BMSError) {
	var equip entity.EquipTab
	if err := m.ds.GetDataSource(ctx, nil).Table(entity.EquipTabName).
		Where("equip_id = ?", equipId).
		First(&equip).GetError(); err != nil {
		return nil, err.Mark()
	}
	return &equip, nil
}

type EquipSearchParam struct {
	SearchKey      string   `json:"search_key"`
	EquipIdList    []string `json:"equip_id_list"`
	EquipName      string   `json:"equip_name"`
	CategoryIdList []int64  `json:"category_id_list"`
	WithLock       bool     `json:"with_lock"`
	PageIn         *paginator.PageIn
}

func (p *EquipManager) SearchEquipMng(ctx context.Context, params *EquipSearchParam) ([]*entity.EquipTab, int64, *bmserror.BMSError) {
	var equipList []*entity.EquipTab
	db := p.ds.GetDataSource(ctx, nil).Table(entity.EquipTabName)
	if params.WithLock {
		db = db.ForUpdate()
	}
	if len(params.CategoryIdList) > 0 {
		db = db.Where("category_id IN (?)", params.CategoryIdList)
	}
	if len(params.EquipIdList) > 0 {
		db = db.Where("equip_id IN (?)", params.EquipIdList)
	}
	if params.EquipName != "" {
		db = db.Where("equip_name = ?", params.EquipName)
	}
	if params.SearchKey != "" {
		db = db.Where("category_name like ? or equip_name like ?", "%"+params.SearchKey+"%")
	}
	total, err := paginator.Paginator(db, params.PageIn, &equipList)
	if err != nil {
		return nil, 0, err.Mark()
	}
	return equipList, total, nil
}
