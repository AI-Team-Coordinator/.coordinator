package activity

import "context"

type triggerKey struct{}

const (
	TriggerManual    = "manual"
	TriggerScheduled = "scheduled"
)

func WithTrigger(ctx context.Context, trigger string) context.Context {
	if trigger == "" {
		return ctx
	}
	return context.WithValue(ctx, triggerKey{}, trigger)
}

func TriggerFrom(ctx context.Context) string {
	if ctx == nil {
		return TriggerScheduled
	}
	v, ok := ctx.Value(triggerKey{}).(string)
	if !ok || v == "" {
		return TriggerScheduled
	}
	return v
}
