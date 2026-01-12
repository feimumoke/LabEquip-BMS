package message

const ApproveBorrowTaskName = "approve_borrow_task"

type ApproveBorrowMessage struct {
	TaskID   string
	Operator string
}

const ReturnBorrowTask = "approve_borrow_task"

type ReturnBorrowMessage struct {
	TaskID   string
	Operator string
}
