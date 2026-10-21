package statebus

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const relayStateGitKeyFile = ".relay_state_git_ssh_key"

// gitExtraEnv returns env vars for git subprocesses (GIT_SSH_COMMAND) when a deploy key is configured.
func gitExtraEnv(dataDir string) ([]string, error) {
	keyPath, err := resolveGitSSHKeyPath(dataDir)
	if err != nil {
		return nil, err
	}
	if keyPath == "" {
		if strings.TrimSpace(os.Getenv("GIT_SSH_COMMAND")) != "" {
			return []string{"GIT_TERMINAL_PROMPT=0"}, nil
		}
		return []string{"GIT_TERMINAL_PROMPT=0"}, nil
	}
	return []string{
		"GIT_TERMINAL_PROMPT=0",
		"GIT_SSH_COMMAND=" + gitSSHCommand(keyPath),
	}, nil
}

func resolveGitSSHKeyPath(dataDir string) (string, error) {
	if p := strings.TrimSpace(os.Getenv("RELAY_STATE_GIT_SSH_KEY_PATH")); p != "" {
		if _, err := os.Stat(p); err != nil {
			return "", fmt.Errorf("RELAY_STATE_GIT_SSH_KEY_PATH: %w", err)
		}
		return p, nil
	}
	pem, err := loadDeployKeyPEMFromEnv()
	if err != nil {
		return "", err
	}
	if len(pem) == 0 {
		return "", nil
	}
	if strings.TrimSpace(dataDir) == "" {
		return "", fmt.Errorf("RELAY_DATA is required to store git deploy key from env")
	}
	path := filepath.Join(dataDir, relayStateGitKeyFile)
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, pem, 0o600); err != nil {
		return "", err
	}
	return path, nil
}

func loadDeployKeyPEMFromEnv() ([]byte, error) {
	if b64 := strings.TrimSpace(os.Getenv("RELAY_STATE_GIT_SSH_KEY_B64")); b64 != "" {
		b64 = strings.TrimRight(b64, "%") // zsh no-newline paste artifact
		b64 = strings.Map(func(r rune) rune {
			if r == '\n' || r == '\r' || r == ' ' {
				return -1
			}
			return r
		}, b64)
		raw, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			return nil, fmt.Errorf("RELAY_STATE_GIT_SSH_KEY_B64: invalid base64")
		}
		if len(raw) == 0 {
			return nil, fmt.Errorf("RELAY_STATE_GIT_SSH_KEY_B64: empty key")
		}
		return raw, nil
	}
	pem := strings.TrimSpace(os.Getenv("RELAY_STATE_GIT_SSH_PRIVATE_KEY"))
	if pem == "" {
		return nil, nil
	}
	if !strings.Contains(pem, "PRIVATE KEY") {
		return nil, fmt.Errorf("RELAY_STATE_GIT_SSH_PRIVATE_KEY: expected PEM private key")
	}
	return []byte(pem + "\n"), nil
}

func gitSSHCommand(keyPath string) string {
	return "ssh -i " + strconv.Quote(keyPath) + " -o IdentitiesOnly=yes -o StrictHostKeyChecking=accept-new"
}
