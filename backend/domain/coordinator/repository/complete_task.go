package repository

import (
	"context"
	"os"
	"os/exec"
)

func (r *FileRepository) CompleteTask(_ context.Context, alias, taskID string) error {
	if alias == "" || taskID == "" {
		return nil
	}
	script := r.appScript("sync_event.sh")
	cmd := exec.Command(script, alias, "task_completed", taskID)
	cmd.Dir = r.appDir()
	cmd.Env = append(os.Environ(), "COORDINATOR_ROOT="+r.appDir())
	return cmd.Run()
}
