package slot

import "time"

// Slot is a process on one person's machine. It is not the task.
// in_progress claims a repo. parked keeps the branch and drops the claim.
type Slot struct {
	Alias     string
	Status    string
	Branch    string
	UpdatedAt time.Time
}
