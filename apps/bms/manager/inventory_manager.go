package manager

import (
	"context"

	"github.com/feimumoke/labequipbms/defines/constant"
	"github.com/feimumoke/labequipbms/defines/entity"
	"github.com/feimumoke/labequipbms/framework/bmserror"
	"github.com/feimumoke/labequipbms/framework/datasource"
	"github.com/feimumoke/labequipbms/framework/support/paginator"
	"github.com/feimumoke/labequipbms/framework/support/timeutil"
)

type InventoryManager struct {
	ds datasource.DataSource
}

func NewInventoryManager() *InventoryManager {
	return &InventoryManager{ds: datasource.DefaultInvSource}
}

// CreateOrUpdateInventory 创建或更新库存，库位唯一约束是lab_id和equip_id
func (m *InventoryManager) CreateOrUpdateInventory(ctx context.Context, labId, equipId string, count int64, operator string) (*entity.InventoryTab, *bmserror.BMSError) {
	var inventory entity.InventoryTab
	db := m.ds.GetDataSource(ctx, nil).Table(entity.InventoryTabName).ForUpdate().
		Where("lab_id = ? AND equip_id = ?", labId, equipId).First(&inventory)

	now := timeutil.GetCurrentUnix()
	if db.RecordNotFound() {
		// 创建新库存
		inventory = entity.InventoryTab{
			LabId:        labId,
			EquipId:      equipId,
			TotalQty:     count,
			AvailableQty: count,
			OnHandQty:    count,
			BorrowedQty:  0,
			AllocatedQty: 0,
			//ReservedQty:  0,
			Operator: operator,
			Ctime:    now,
			Mtime:    now,
		}
		if err := m.ds.GetDataSource(ctx, nil).Table(entity.InventoryTabName).Create(&inventory).GetError(); err != nil {
			return nil, bmserror.NewError(constant.ErrDB, err.Error())
		}
	} else {
		if err := db.GetError(); err != nil {
			return nil, bmserror.NewError(constant.ErrDB, err.Error())
		}
		// 更新库存
		inventory.TotalQty += count
		inventory.AvailableQty += count
		inventory.OnHandQty += count
		inventory.Operator = operator
		inventory.Mtime = now

		updateMap := map[string]interface{}{
			"total_qty":     inventory.TotalQty,
			"available_qty": inventory.AvailableQty,
			"on_hand_qty":   inventory.OnHandQty,
			"operator":      operator,
			"mtime":         now,
		}
		if err := m.ds.GetDataSource(ctx, nil).Table(entity.InventoryTabName).
			Where("lab_id = ? AND equip_id = ?", labId, equipId).
			Updates(updateMap).GetError(); err != nil {
			return nil, bmserror.NewError(constant.ErrDB, err.Error())
		}
	}
	return &inventory, nil
}

// GetInventoryByLabAndEquip 根据lab_id和equip_id获取库存
func (m *InventoryManager) GetInventoryByLabAndEquip(ctx context.Context, labId, equipId string, withLock bool) (*entity.InventoryTab, *bmserror.BMSError) {
	var inventory entity.InventoryTab
	db := m.ds.GetDataSource(ctx, nil).Table(entity.InventoryTabName).
		Where("lab_id = ? AND equip_id = ?", labId, equipId)
	if withLock {
		db = db.ForUpdate()
	}
	if err := db.First(&inventory).GetError(); err != nil {
		if db.RecordNotFound() {
			return nil, nil
		}
		return nil, bmserror.NewError(constant.ErrDB, err.Error())
	}
	return &inventory, nil
}

// UpdateInventory 更新库存
func (m *InventoryManager) UpdateInventory(ctx context.Context, inventory *entity.InventoryTab) *bmserror.BMSError {
	now := timeutil.GetCurrentUnix()
	updateMap := map[string]interface{}{
		"total_qty":     inventory.TotalQty,
		"available_qty": inventory.AvailableQty,
		"borrowed_qty":  inventory.BorrowedQty,
		"on_hand_qty":   inventory.OnHandQty,
		"allocated_qty": inventory.AllocatedQty,
		"operator":      inventory.Operator,
		"mtime":         now,
	}
	if err := m.ds.GetDataSource(ctx, nil).Table(entity.InventoryTabName).
		Where("id = ? AND lab_id = ? AND equip_id = ?", inventory.ID, inventory.LabId, inventory.EquipId).
		Updates(updateMap).GetError(); err != nil {
		return bmserror.NewError(constant.ErrDB, err.Error())
	}
	inventory.Mtime = now
	return nil
}

// SearchInventory 查询库存
type InventorySearchParam struct {
	LabCode string
	EquipId string
	PageIn  *paginator.PageIn
}

func (m *InventoryManager) SearchInventory(ctx context.Context, params *InventorySearchParam) ([]*entity.InventoryTab, int64, *bmserror.BMSError) {
	var inventoryList []*entity.InventoryTab
	db := m.ds.GetDataSource(ctx, nil).Table(entity.InventoryTabName)
	if params.LabCode != "" {
		db = db.Where("lab_id = ?", params.LabCode)
	}
	if params.EquipId != "" {
		db = db.Where("equip_id = ?", params.EquipId)
	}
	total, err := paginator.Paginator(db, params.PageIn, &inventoryList)
	if err != nil {
		return nil, 0, err.Mark()
	}
	return inventoryList, total, nil
}

// DecreaseInventory 扣减库存
func (m *InventoryManager) DecreaseInventory(ctx context.Context, labId, equipId string, count int64, operator string) *bmserror.BMSError {
	inventory, bmsErr := m.GetInventoryByLabAndEquip(ctx, labId, equipId, true)
	if bmsErr != nil {
		return bmsErr.Mark()
	}
	if inventory == nil {
		return bmserror.NewError(constant.ErrParam, "inventory not found")
	}
	if inventory.AvailableQty < count {
		return bmserror.NewError(constant.ErrParam, "available quantity is not enough")
	}
	inventory.TotalQty -= count
	inventory.OnHandQty -= count
	inventory.AvailableQty -= count
	inventory.Operator = operator
	inventory.Mtime = timeutil.GetCurrentUnix()
	return m.UpdateInventory(ctx, inventory)
}

// BatchGetInventoryByLabAndEquip 批量获取库存
func (m *InventoryManager) BatchGetInventoryByLabAndEquip(ctx context.Context, labEquipPairs []struct{ LabId, EquipId string }) (map[string]*entity.InventoryTab, *bmserror.BMSError) {
	if len(labEquipPairs) == 0 {
		return make(map[string]*entity.InventoryTab), nil
	}

	var inventoryList []*entity.InventoryTab
	db := m.ds.GetDataSource(ctx, nil).Table(entity.InventoryTabName)

	// 构建查询条件：使用 OR 连接多个 (lab_id, equip_id) 对
	query := db
	for i, pair := range labEquipPairs {
		if i == 0 {
			query = query.Where("(lab_id = ? AND equip_id = ?)", pair.LabId, pair.EquipId)
		} else {
			query = query.Or("(lab_id = ? AND equip_id = ?)", pair.LabId, pair.EquipId)
		}
	}

	if err := query.Find(&inventoryList).GetError(); err != nil {
		return nil, bmserror.NewError(constant.ErrDB, err.Error())
	}

	inventoryMap := make(map[string]*entity.InventoryTab)
	for _, inv := range inventoryList {
		key := inv.LabId + "_" + inv.EquipId
		inventoryMap[key] = inv
	}
	return inventoryMap, nil
}
