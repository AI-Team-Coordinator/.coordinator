package model

// External event class: how Coordinator should treat the signal (Pulse, Otto).
// Adapters pick a known class; the set grows without Coordinator releases.
const (
	ClassInfo     = "info"
	ClassWarning  = "warning"
	ClassReminder = "reminder"
)

func ValidClass(c string) bool {
	switch c {
	case ClassInfo, ClassWarning, ClassReminder:
		return true
	default:
		return false
	}
}
