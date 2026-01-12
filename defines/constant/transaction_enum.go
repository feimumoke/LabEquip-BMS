package constant

// TransactionSheetType 交易单据类型
type TransactionSheetType = int64

const (
	TransactionSheetTypeInventory TransactionSheetType = 1 //库存单
	TransactionSheetTypeBorrow    TransactionSheetType = 2 //借记单
)

// 导出给前端的交易单据类型枚举
var ExportTransactionSheetTypeNameToValue = map[string]interface{}{
	"库存单": TransactionSheetTypeInventory,
	"借记单": TransactionSheetTypeBorrow,
}

// TransactionType 交易类型
type TransactionType = int64

const (
	TransactionTypeIncrease TransactionType = 1 //增加库存
	TransactionTypeDecrease TransactionType = 2 //扣减库存

	TransactionTypeAllocate TransactionType = 3 //分配库存
	TransactionTypeBorrow   TransactionType = 4 //借出设备
	TransactionTypeReturn   TransactionType = 5 //归还设备
	TransactionTypeReject   TransactionType = 6 //拒绝分配
)

// 导出给前端的交易类型枚举
var ExportTransactionTypeNameToValue = map[string]interface{}{
	"增加库存": TransactionTypeIncrease,
	"扣减库存": TransactionTypeDecrease,
	"分配库存": TransactionTypeAllocate,
	"借出设备": TransactionTypeBorrow,
	"归还设备": TransactionTypeReturn,
	"拒绝分配": TransactionTypeReject,
}

func init() {
	RegisterEnumValues("TransactionType", ExportTransactionTypeNameToValue)
	RegisterEnumValues("TransactionSheetType", ExportTransactionSheetTypeNameToValue)
}
