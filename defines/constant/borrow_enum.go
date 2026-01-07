package constant

type BorrowTaskStatus = int64

const (
	BorrowTaskStatusPending  BorrowTaskStatus = 1 //创建了 没分配库存
	BorrowTaskStatusAllocate BorrowTaskStatus = 2 //已经分配库存
	BorrowTaskStatusApproval BorrowTaskStatus = 3 //已经审批通过
	BorrowTaskStatusOngoing  BorrowTaskStatus = 4 //已经拿走了

	BorrowTaskStatusReject BorrowTaskStatus = 7 //审批拒绝
	BorrowTaskStatusDone   BorrowTaskStatus = 8 //归还
	BorrowTaskStatusCancel BorrowTaskStatus = 9 //取消
)

var DoneBorrowTaskStatusList = []BorrowTaskStatus{BorrowTaskStatusReject, BorrowTaskStatusDone, BorrowTaskStatusCancel}
