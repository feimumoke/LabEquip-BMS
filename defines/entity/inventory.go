package entity

type InventoryTab struct {
	ID              int64  `gorm:"column:id;primary_key" json:"id"`
	LabId           string `gorm:"column:lab_id" json:"lab_id"`
	EquipId         string `gorm:"column:equip_id" json:"equip_id"`
	TotalQty        int64  `gorm:"column:total_qty" json:"total_qty"`
	AvailableQty    int64  `gorm:"column:available_qty" json:"available_qty"`         //锁定库存
	BorrowedQty     int64  `gorm:"column:borrowed_qty" json:"borrowed_qty"`           //当前被用户实际借走的
	PreAllocatedQty int64  `gorm:"column:pre_allocated_qty" json:"pre_allocated_qty"` //已经分配
	ReservedQty     int64  `gorm:"column:reserved_qty" json:"reserved_qty"`           //已被预约，但尚未借出的，先分配
	Operator        string `gorm:"column:operator" json:"operator"`
	Ctime           int64  `gorm:"column:ctime" json:"ctime"`
	Mtime           int64  `gorm:"column:mtime" json:"mtime"`
}
