package completion

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectShell_Override(t *testing.T) {
	sh, err := DetectShell("bash")
	require.NoError(t, err)
	assert.Equal(t, Bash, sh)

	sh, err = DetectShell("zsh")
	require.NoError(t, err)
	assert.Equal(t, Zsh, sh)

	sh, err = DetectShell("fish")
	require.NoError(t, err)
	assert.Equal(t, Fish, sh)
}

func TestDetectShell_FullPath(t *testing.T) {
	sh, err := DetectShell("/usr/bin/bash")
	require.NoError(t, err)
	assert.Equal(t, Bash, sh)
}

func TestDetectShell_FromEnv(t *testing.T) {
	t.Setenv("SHELL", "/bin/zsh")
	sh, err := DetectShell("")
	require.NoError(t, err)
	assert.Equal(t, Zsh, sh)
}

func TestDetectShell_Unsupported(t *testing.T) {
	_, err := DetectShell("tcsh")
	assert.ErrorContains(t, err, "unsupported shell")
}

func TestDetectShell_Empty(t *testing.T) {
	t.Setenv("SHELL", "")
	_, err := DetectShell("")
	assert.ErrorContains(t, err, "could not detect shell")
}

func testRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cwai",
		Short: "test",
	}
}

func TestInstall_Bash(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_DATA_HOME", "")

	sh, path, err := Install(testRootCmd(), "bash")
	require.NoError(t, err)
	assert.Equal(t, Bash, sh)
	assert.Equal(t, filepath.Join(tmpDir, ".local", "share", "bash-completion", "completions", "cwai"), path)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), marker)
}

func TestInstall_Zsh(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	sh, path, err := Install(testRootCmd(), "zsh")
	require.NoError(t, err)
	assert.Equal(t, Zsh, sh)
	assert.Equal(t, filepath.Join(tmpDir, ".zsh", "completions", "_cwai"), path)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), marker)
}

func TestInstall_Fish(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, ".config"))

	sh, path, err := Install(testRootCmd(), "fish")
	require.NoError(t, err)
	assert.Equal(t, Fish, sh)
	assert.Equal(t, filepath.Join(tmpDir, ".config", "fish", "completions", "cwai.fish"), path)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), marker)
}

func TestInstall_AlreadyInstalled(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_DATA_HOME", "")

	_, _, err := Install(testRootCmd(), "bash")
	require.NoError(t, err)

	_, _, err = Install(testRootCmd(), "bash")
	assert.ErrorContains(t, err, "already installed")
}

func TestInstall_ForeignFileExists(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_DATA_HOME", "")

	path := filepath.Join(tmpDir, ".local", "share", "bash-completion", "completions", "cwai")
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte("foreign content"), 0o644))

	_, _, err := Install(testRootCmd(), "bash")
	assert.ErrorContains(t, err, "not managed by cwai")
}

func TestUninstall_Bash(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_DATA_HOME", "")

	_, _, err := Install(testRootCmd(), "bash")
	require.NoError(t, err)

	sh, path, err := Uninstall("bash")
	require.NoError(t, err)
	assert.Equal(t, Bash, sh)
	assert.Equal(t, filepath.Join(tmpDir, ".local", "share", "bash-completion", "completions", "cwai"), path)

	_, err = os.Stat(path)
	assert.True(t, os.IsNotExist(err))
}

func TestUninstall_NotInstalled(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_DATA_HOME", "")

	_, _, err := Uninstall("bash")
	assert.ErrorContains(t, err, "no cwai completion found")
}

func TestUninstall_ForeignFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_DATA_HOME", "")

	path := filepath.Join(tmpDir, ".local", "share", "bash-completion", "completions", "cwai")
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte("foreign content"), 0o644))

	_, _, err := Uninstall("bash")
	assert.ErrorContains(t, err, "not installed by cwai")
}

func TestInstall_Bash_XDGDataHome(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir, "custom-data"))

	sh, path, err := Install(testRootCmd(), "bash")
	require.NoError(t, err)
	assert.Equal(t, Bash, sh)
	assert.Equal(t, filepath.Join(tmpDir, "custom-data", "bash-completion", "completions", "cwai"), path)
}

func TestPostInstallHint(t *testing.T) {
	assert.Contains(t, PostInstallHint(Bash), "automatically")
	assert.Contains(t, PostInstallHint(Zsh), "zshrc")
	assert.Contains(t, PostInstallHint(Fish), "automatically")
}

func TestInstall_Fish_XDGDefault(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", "")

	sh, path, err := Install(testRootCmd(), "fish")
	require.NoError(t, err)
	assert.Equal(t, Fish, sh)
	assert.Equal(t, filepath.Join(tmpDir, ".config", "fish", "completions", "cwai.fish"), path)
}
