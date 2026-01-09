package entity

const InventoryTabName = "inventory_tab"

type InventoryTab struct {
	ID           int64  `gorm:"column:id;primary_key" json:"id"`
	LabId        string `gorm:"column:lab_id" json:"lab_id"`
	EquipId      string `gorm:"column:equip_id" json:"equip_id"`
	TotalQty     int64  `gorm:"column:total_qty" json:"total_qty"`
	OnHandQty    int64  `gorm:"column:on_hand_qty" json:"on_hand_qty"`     //在实验室库存 = total - borrow
	AvailableQty int64  `gorm:"column:available_qty" json:"available_qty"` //可用库存 = total -borrow - prev allocate
	BorrowedQty  int64  `gorm:"column:borrowed_qty" json:"borrowed_qty"`   //当前被用户实际借走的[已经拿走了] - 可用库存不变
	AllocatedQty int64  `gorm:"column:allocated_qty" json:"allocated_qty"` //已经分配 - 可用库存会减少 allocated
	//ReservedQty  int64  `gorm:"column:reserved_qty" json:"reserved_qty"`   //已被预约，pending 状态的
	Operator string `gorm:"column:operator" json:"operator"`
	Ctime    int64  `gorm:"column:ctime" json:"ctime"`
	Mtime    int64  `gorm:"column:mtime" json:"mtime"`
}
