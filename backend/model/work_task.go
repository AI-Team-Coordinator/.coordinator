package model

import (
	"coordinator/domain/slot"
	"coordinator/domain/task"
)

// WorkTask is the snapshot name for a task. The type lives in domain/task.
type WorkTask = task.Task

// WorkTaskSlot is the snapshot name for a slot linked to a task. The type lives in domain/slot.
type WorkTaskSlot = slot.Slot
