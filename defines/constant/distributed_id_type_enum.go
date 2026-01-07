package constant

type DistributedIDType = int64

const (
	EquipID        = 101
	EquipInvTaskId = 151
	BorrowTaskID   = 201
	TransactionID  = 301
	UserNo         = 400
)

var DistributedIDTypePerfixMap = map[DistributedIDType]string{
	TransactionID:  "TS",
	EquipID:        "EP",
	EquipInvTaskId: "IV",
	BorrowTaskID:   "BR",
}

var IDTypeToStep = map[DistributedIDType]int64{
	TransactionID:  20,
	EquipID:        1,
	EquipInvTaskId: 1,
	BorrowTaskID:   10,
	UserNo:         1,
}
