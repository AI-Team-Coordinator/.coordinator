package repository

import (
	"context"

	"coordinator/model"
)

// CoordinatorRepository is the data port for team state and history.
// FileRepository is the local-first implementation. A cloud source can
// implement the same interface later without changing service or HTTP.
type CoordinatorRepository interface {
	GetAuthors(ctx context.Context) (map[string]string, error)
	GetTeam(ctx context.Context) ([]model.TeamPerson, error)
	SaveTeam(ctx context.Context, members []model.TeamPerson) error
	GetProject(ctx context.Context) (*model.ProjectProfile, error)
	SaveProject(ctx context.Context, profile *model.ProjectProfile) error
	WorkspaceDir() string
	GetMembers(ctx context.Context) ([]model.Member, error)
	CursorUsageSnapshot() *model.CursorUsage
	LoadUsageSamples() []model.UsageSample
	CompleteTask(ctx context.Context, alias, taskID string) error
	GetStrayWork(ctx context.Context, members []model.Member) ([]model.StrayRepo, error)
	CurrentAuthor(ctx context.Context) (string, error)
	SettingsSyncStatus(ctx context.Context) (*model.SyncStatus, error)
	SyncSettings(ctx context.Context, alias string) (*model.SyncResult, error)
	GetTaskDoc(ctx context.Context, taskID string) (*model.TaskDoc, error)
	TaskTitle(ctx context.Context, taskID string) string
	AppendEvent(ctx context.Context, ev model.Event) error
}
