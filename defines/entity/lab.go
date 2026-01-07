package entity

const LabTabName = "lab_tab"

type LaboratoryTab struct {
	ID          int64  `gorm:"column:id;primary_key" json:"id" comment:"主键"`
	LabCode     string `gorm:"column:lab_code" json:"lab_code"`
	LabName     string `gorm:"column:lab_name" json:"lab_name"`
	Address     string `gorm:"column:address" json:"address"`
	ManagerId   int64  `gorm:"column:manager_id" json:"manager_id"` //管理员
	Description string `gorm:"column:description" json:"description"`
	Ctime       int64  `gorm:"column:ctime" json:"ctime" comment:"时间戳"`
	Mtime       int64  `gorm:"column:mtime" json:"mtime" comment:"时间戳"`
}
