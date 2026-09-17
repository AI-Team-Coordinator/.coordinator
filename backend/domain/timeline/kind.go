// Package timeline is the append-only log of coordination facts.
// jsonl (model.Event) is the store; Handoff and Decision are typed payloads.
package timeline

const (
	KindTaskStarted       = "task_started"
	KindTaskCompleted     = "task_completed"
	KindTaskParked        = "task_parked"
	KindResearchStarted   = "research_started"
	KindResearchCompleted = "research_completed"
	KindRepoMerged        = "repo_merged"
	KindDeployStarted     = "deploy_started"
	KindDeployFinished    = "deploy_finished"
	KindDeployFailed      = "deploy_failed"
	KindWarning           = "coordinator_warning"
	KindStop              = "coordinator_stop"
	KindHandoff           = "handoff"
	KindDecision          = "decision"
)
