package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"coordinator/model"
)

func (r *FileRepository) eventsFile(alias string) string {
	year := strconv.Itoa(time.Now().Year())
	return filepath.Join(r.progressDir(), "events", alias, year+".jsonl")
}

func (r *FileRepository) AppendEvent(_ context.Context, ev model.Event) error {
	alias := ev.Alias
	if alias == "" {
		return fmt.Errorf("event alias is empty")
	}
	if ev.Event == "" {
		return fmt.Errorf("event type is empty")
	}
	if ev.Timestamp == 0 {
		ev.Timestamp = time.Now().Unix()
	}

	r.gitMu.Lock()
	defer r.gitMu.Unlock()

	path := r.eventsFile(alias)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	line, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	_, writeErr := f.Write(append(line, '\n'))
	closeErr := f.Close()
	if writeErr != nil {
		return writeErr
	}
	if closeErr != nil {
		return closeErr
	}
	r.pushCoordinatorState(alias, ev.Event)
	return nil
}

func (r *FileRepository) pushCoordinatorState(alias, event string) {
	script := r.appScript("coordinator_state.sh")
	if !fileExists(script) {
		return
	}
	appDir := r.appDir()
	msg := fmt.Sprintf("chore(progress): %s %s", alias, event)
	go func() {
		cmd := exec.Command(script, "push", msg)
		cmd.Dir = appDir
		cmd.Env = append(os.Environ(), "COORDINATOR_ROOT="+appDir)
		_ = cmd.Run()
	}()
}
