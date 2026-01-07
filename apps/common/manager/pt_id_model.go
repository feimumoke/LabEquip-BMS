package cmanager

const DistributedIDcreatorTabTableName = "distributed_idcreator_tab"

type DistributedIDcreatorTab struct {
	ID          int64  `gorm:"column:id;primary_key" json:"id"`
	IDType      int64  `gorm:"column:id_type" json:"id_type"`
	DatePeriod  int64  `gorm:"column:date_period" json:"date_period"`
	PtNo        int64  `gorm:"column:pt_no" json:"pt_no"`
	IDValue     int64  `gorm:"column:id_value" json:"id_value"`
	Description string `gorm:"column:description" json:"description"`
	Mtime       int64  `gorm:"column:mtime" json:"mtime"`
}
