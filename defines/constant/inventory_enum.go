package constant

type InventoryTaskType = int64

const (
	InventoryTaskTypeIncrease InventoryTaskType = 1
	InventoryTaskTypeDecrease InventoryTaskType = 2
	InventoryTaskTypeTransfer InventoryTaskType = 3
)
