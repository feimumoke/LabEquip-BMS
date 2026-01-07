package constant

type EquipCategory = int64

const (
	EquipCategoryGeneral     EquipCategory = 1 //通用实验设备
	EquipCategoryChemical    EquipCategory = 2 //化学实验设备
	EquipCategoryBiological  EquipCategory = 3 //生物实验设备
	EquipCategoryPhysics     EquipCategory = 4 //物理实验设备
	EquipCategoryElectronics EquipCategory = 5 //电子实验设备
	EquipCategoryIT          EquipCategory = 6 //计算机/信息设备
	EquipCategorySafety      EquipCategory = 7 //安全与防护设备
	EquipCategorySpecial     EquipCategory = 8 //特殊教学设备
)

var ExportCategoryNameToValue = map[string]interface{}{
	"通用实验设备":   EquipCategoryGeneral,
	"化学实验设备":   EquipCategoryChemical,
	"生物实验设备":   EquipCategoryBiological,
	"物理实验设备":   EquipCategoryPhysics,
	"电子实验设备":   EquipCategoryElectronics,
	"计算机/信息设备": EquipCategoryIT,
	"安全与防护设备":  EquipCategorySafety,
	"特殊教学设备":   EquipCategorySpecial,
}

func GetCategoryName(idType EquipCategory) string {
	for k, v := range ExportCategoryNameToValue {
		if v == idType {
			return k
		}
	}
	return ""
}

func init() {
	RegisterEnumValues("EquipCategory", ExportCategoryNameToValue)
}
