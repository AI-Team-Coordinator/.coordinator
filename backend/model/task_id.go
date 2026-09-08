package model

import (
	"regexp"
	"strings"
)

var taskIDPattern = regexp.MustCompile(`(?i)\b(?:FIX-)?\d{8}-\d{4}-[A-Z]{2,4}-[A-Z0-9_]+\b`)
var validTaskID = regexp.MustCompile(`^(?i)(?:FIX-)?\d{8}-\d{4}-[A-Z]{2,4}-[A-Z0-9_]+$`)

func ValidTaskID(taskID string) bool {
	return validTaskID.MatchString(strings.TrimSpace(taskID))
}

// MessageHasTaskID reports whether a commit subject/body mentions a coordinator task id.
func MessageHasTaskID(message string) bool {
	return taskIDPattern.FindString(message) != ""
}

// MessageHasThisTask reports whether the message refers to this task_id.
func MessageHasThisTask(message, taskID string) bool {
	if taskID == "" {
		return false
	}
	return regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(taskID) + `\b`).MatchString(message)
}
