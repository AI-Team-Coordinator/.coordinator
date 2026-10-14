package task

// Status is backlog when nobody has a slot, open while a slot exists, done after completion.
func Status(stored string, slotCount int) string {
	if stored == "done" {
		return "done"
	}
	if slotCount > 0 {
		return "open"
	}
	if stored == "open" {
		return "backlog"
	}
	return stored
}
