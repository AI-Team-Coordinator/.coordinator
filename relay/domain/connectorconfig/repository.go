package connectorconfig

import "context"

// Repository persists connector meta on the relay instance database.
type Repository interface {
	Get(ctx context.Context, connectorID string) (Meta, bool, error)
	SetEnabled(ctx context.Context, connectorID string, on bool) error
	MergeNonSecret(ctx context.Context, connectorID string, patch map[string]any) error
}
