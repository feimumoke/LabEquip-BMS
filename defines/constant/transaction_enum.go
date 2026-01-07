package constant

type TransactionSheetType = int64

const (
	TransactionSheetTypeInventory TransactionSheetType = 1
	TransactionSheetTypeBorrow    InventoryTaskType    = 2
)

type TransactionType = int64

const (
	TransactionTypeIncrease TransactionType = 1
	TransactionTypeDecrease TransactionType = 2
)
