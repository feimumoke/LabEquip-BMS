package entity

import "github.com/feimumoke/labequipbms/defines/constant"

const InventoryTaskTabName = "inventory_task_tab"

type InventoryTaskTab struct {
	ID         int64                        `gorm:"column:id;primary_key" json:"id"`
	TaskID     string                       `gorm:"column:task_id" json:"task_id"`
	TaskType   constant.InventoryTaskType   `gorm:"column:task_type" json:"task_type"`
	TaskStatus constant.InventoryTaskStatus `gorm:"column:task_status" json:"task_status"`
	LabId      string                       `gorm:"column:lab_id" json:"lab_id"`
	EquipId    string                       `gorm:"column:equip_id" json:"equip_id"`
	TotalQty   int64                        `gorm:"column:total_qty" json:"total_qty"` //已被预约，但尚未借出的，先分配
	Operator   string                       `gorm:"column:operator" json:"operator"`
	Remark     string                       `gorm:"column:remark" json:"remark"`
	Ctime      int64                        `gorm:"column:ctime" json:"ctime"`
	Mtime      int64                        `gorm:"column:mtime" json:"mtime"`
}

const InventoryTaskLogTabName = "inventory_task_log_tab"

type InventoryTaskLogTab struct {
	ID         int64                        `gorm:"column:id;primary_key" json:"id" comment:"主键"`
	TaskID     string                       `gorm:"column:task_id" json:"task_id" comment:"任务id"`
	TaskStatus constant.InventoryTaskStatus `gorm:"column:task_status" json:"task_status"`
	Remark     string                       `gorm:"column:remark" json:"remark"`
	Operator   string                       `gorm:"column:operator" json:"operator"`
}
