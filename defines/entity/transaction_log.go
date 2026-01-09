package entity

import "github.com/feimumoke/labequipbms/defines/constant"

const TransactionLogTabName = "transaction_log_tab"

type TransactionLogTab struct {
	ID            int64                         `gorm:"column:id;primary_key" json:"id" comment:"主键"`
	TransactionID string                        `gorm:"column:transaction_id" json:"transaction_id" comment:"交易ID"`
	SheetID       string                        `gorm:"column:sheet_id" json:"sheet_id" comment:"单据ID"`
	EquipId       string                        `gorm:"column:equip_id" json:"equip_id" comment:"设备ID"`
	LabId         string                        `gorm:"column:lab_id" json:"lab_id" comment:"实验室ID"`
	Operator      string                        `gorm:"column:operator" json:"operator" comment:"操作人"`
	Remark        string                        `gorm:"column:remark" json:"remark" comment:"备注"`
	TotalQty      int64                         `gorm:"column:total_qty" json:"total_qty"`
	OnHandQty     int64                         `gorm:"column:on_hand_qty" json:"on_hand_qty"`     //在实验室库存 = total - borrow
	AvailableQty  int64                         `gorm:"column:available_qty" json:"available_qty"` //可用库存 = total -borrow - prev allocate
	BorrowedQty   int64                         `gorm:"column:borrowed_qty" json:"borrowed_qty"`   //当前被用户实际借走的[已经拿走了] - 可用库存不变
	AllocatedQty  int64                         `gorm:"column:allocated_qty" json:"allocated_qty"` //已经分配 - 可用库存会减少 allocated
	OpQty         int64                         `gorm:"column:op_qty" json:"op_qty" comment:"操作数量"`
	TransType     constant.TransactionType      `gorm:"column:trans_type" json:"trans_type" comment:"交易类型"`
	SheetType     constant.TransactionSheetType `gorm:"column:sheet_type" json:"sheet_type" comment:"单据类型"`
	Ctime         int64                         `gorm:"column:ctime" json:"ctime" comment:"时间戳"`
}
