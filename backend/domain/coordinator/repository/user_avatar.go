package repository

import (
	"os"
	"path/filepath"
	"strings"
)

func (r *FileRepository) avatarsDir() string {
	return filepath.Join(r.dataDir(), "avatars")
}

func (r *FileRepository) userAvatarPath(alias string) string {
	safe := strings.ToUpper(strings.TrimSpace(alias))
	return filepath.Join(r.avatarsDir(), safe+".jpg")
}

func (r *FileRepository) UserAvatarExists(alias string) bool {
	if strings.TrimSpace(alias) == "" {
		return false
	}
	_, err := os.Stat(r.userAvatarPath(alias))
	return err == nil
}

func (r *FileRepository) ReadUserAvatar(alias string) ([]byte, error) {
	return os.ReadFile(r.userAvatarPath(alias))
}

func (r *FileRepository) WriteUserAvatar(alias string, data []byte) error {
	if err := os.MkdirAll(r.avatarsDir(), 0o755); err != nil {
		return err
	}
	return writeFileAtomic(r.userAvatarPath(alias), data)
}

func (r *FileRepository) DeleteUserAvatar(alias string) error {
	if err := os.Remove(r.userAvatarPath(alias)); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
