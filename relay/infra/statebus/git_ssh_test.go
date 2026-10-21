package statebus

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadDeployKeyPEMFromEnvB64(t *testing.T) {
	pem := "-----BEGIN OPENSSH PRIVATE KEY-----\nabc\n-----END OPENSSH PRIVATE KEY-----\n"
	t.Setenv("RELAY_STATE_GIT_SSH_PRIVATE_KEY", "")
	t.Setenv("RELAY_STATE_GIT_SSH_KEY_B64", base64.StdEncoding.EncodeToString([]byte(pem)))
	got, err := loadDeployKeyPEMFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != pem {
		t.Fatalf("pem mismatch")
	}
}

func TestResolveGitSSHKeyPathWritesFile(t *testing.T) {
	pem := "-----BEGIN OPENSSH PRIVATE KEY-----\nxyz\n-----END OPENSSH PRIVATE KEY-----\n"
	t.Setenv("RELAY_STATE_GIT_SSH_KEY_PATH", "")
	t.Setenv("RELAY_STATE_GIT_SSH_KEY_B64", base64.StdEncoding.EncodeToString([]byte(pem)))
	dir := t.TempDir()
	path, err := resolveGitSSHKeyPath(dir)
	if err != nil {
		t.Fatal(err)
	}
	if path != filepath.Join(dir, relayStateGitKeyFile) {
		t.Fatalf("path %s", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != pem {
		t.Fatal("file content")
	}
}

func TestGitSSHCommandQuotesPath(t *testing.T) {
	cmd := gitSSHCommand("/data/key with spaces")
	if !strings.Contains(cmd, "\"/data/key with spaces\"") {
		t.Fatalf("cmd %q", cmd)
	}
}
