package constant

type BorrowTaskStatus = int64

const (
	BorrowTaskStatusPending  BorrowTaskStatus = 1 //创建了 没分配库存
	BorrowTaskStatusApproval BorrowTaskStatus = 2 //已经审批通过
	BorrowTaskStatusAllocate BorrowTaskStatus = 3 //已经分配库存
	BorrowTaskStatusOngoing  BorrowTaskStatus = 4 //已经拿走了

	BorrowTaskStatusReject BorrowTaskStatus = 7 //审批拒绝
	BorrowTaskStatusDone   BorrowTaskStatus = 8 //归还
	BorrowTaskStatusCancel BorrowTaskStatus = 9 //取消
)

var DoneBorrowTaskStatusList = []BorrowTaskStatus{BorrowTaskStatusReject, BorrowTaskStatusDone, BorrowTaskStatusCancel}

// 导出给前端的借记任务状态枚举
var ExportBorrowTaskStatusNameToValue = map[string]interface{}{
	"待审批": BorrowTaskStatusPending,
	"已审批": BorrowTaskStatusApproval,
	"已分配": BorrowTaskStatusAllocate,
	"进行中": BorrowTaskStatusOngoing,
	"已拒绝": BorrowTaskStatusReject,
	"已归还": BorrowTaskStatusDone,
	"已取消": BorrowTaskStatusCancel,
}

func init() {
	RegisterEnumValues("BorrowTaskStatus", ExportBorrowTaskStatusNameToValue)
}
