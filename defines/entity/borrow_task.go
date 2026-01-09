package entity

import (
	"github.com/feimumoke/labequipbms/defines/constant"
	"github.com/feimumoke/labequipbms/framework/support/collection"
)

const BorrowTaskTabName = "borrow_task_tab"

type BorrowTask struct {
	ID         int64                     `gorm:"column:id;primary_key" json:"id" comment:"主键"`
	TaskID     string                    `gorm:"column:task_id" json:"task_id" comment:"任务id"`
	EquipId    string                    `gorm:"column:equip_id" json:"equip_id"`
	LabId      string                    `gorm:"column:lab_id" json:"lab_id"`
	BorrowQty  int64                     `gorm:"column:borrow_qty" json:"borrow_qty" comment:"借记数量"`
	TaskStatus constant.BorrowTaskStatus `gorm:"column:task_status" json:"task_status"`
	Creator    string                    `gorm:"column:creator" json:"creator"`
	Approval   string                    `gorm:"column:approval" json:"approval"`
	Ctime      int64                     `gorm:"column:ctime" json:"ctime" comment:"时间戳"`
	Mtime      int64                     `gorm:"column:mtime" json:"mtime" comment:"时间戳"`
}

func (t *BorrowTask) IsDone() bool {
	return collection.Contain(t.TaskStatus, constant.DoneBorrowTaskStatusList)
}

const BorrowTaskLogTabName = "borrow_task_log_tab"

type BorrowTaskLog struct {
	ID         int64                     `gorm:"column:id;primary_key" json:"id" comment:"主键"`
	TaskID     string                    `gorm:"column:task_id" json:"task_id" comment:"任务id"`
	TaskStatus constant.BorrowTaskStatus `gorm:"column:task_status" json:"task_status"`
	Remark     string                    `gorm:"column:remark" json:"remark"`
	Operator   string                    `gorm:"column:operator" json:"operator"`
	Ctime      int64                     `gorm:"column:ctime" json:"ctime" comment:"时间戳"`
}
