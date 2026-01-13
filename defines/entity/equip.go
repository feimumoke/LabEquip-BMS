package entity

import (
	"strings"

	"github.com/feimumoke/labequipbms/defines/constant"
)

const EquipTabName = "equip_tab"

type EquipTab struct {
	ID           int64                  `gorm:"column:id;primary_key" json:"id" comment:"主键"`
	EquipId      string                 `gorm:"column:equip_id" json:"equip_id"`
	CategoryId   constant.EquipCategory `gorm:"column:category_id" json:"category_id"`
	CategoryName string                 `gorm:"column:category_name" json:"category_name"`
	EquipName    string                 `gorm:"column:equip_name" json:"equip_name"`
	Model        string                 `gorm:"column:model" json:"model"`
	Images       string                 `gorm:"column:images" json:"images"`
	Creator      string                 `gorm:"column:creator" json:"creator"`
	Description  string                 `gorm:"column:description" json:"description"`
	Ctime        int64                  `gorm:"column:ctime" json:"ctime" comment:"时间戳"`
	Mtime        int64                  `gorm:"column:mtime" json:"mtime" comment:"时间戳"`
}

func (e EquipTab) GetImages() []string {
	if e.Images == "" {
		return []string{}
	}
	return strings.Split(e.Images, ",")
}

func (e *EquipTab) GetImageUrlList() []string {
	images := e.GetImages()
	urlList := make([]string, 0)
	for _, image := range images {
		urlList = append(urlList, constant.EquipSystemDomain+image)
	}
	return urlList
}
