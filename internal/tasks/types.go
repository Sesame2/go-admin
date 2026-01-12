package tasks

type TaskType string

const (
	DocumentProcess TaskType = "document_process"
)

type TaskStatus string

const (
	// 待定状态
	StatusPending TaskStatus = "pending"

	// 处理中状态
	StatusProcessing TaskStatus = "processing"

	// 处理成功状态
	StatusCompleted TaskStatus = "completed"

	// 失败状态
	StatusFailed TaskStatus = "failed"
)
