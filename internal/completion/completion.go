package completion

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

const marker = "# Installed by: cwai completion install"

type Shell string

const (
	Bash Shell = "bash"
	Zsh  Shell = "zsh"
	Fish Shell = "fish"
)

func DetectShell(override string) (Shell, error) {
	name := override
	if name == "" {
		name = os.Getenv("SHELL")
	}
	if name == "" {
		return "", fmt.Errorf("could not detect shell: $SHELL is not set. Use --shell flag")
	}

	base := filepath.Base(name)
	switch base {
	case "bash":
		return Bash, nil
	case "zsh":
		return Zsh, nil
	case "fish":
		return Fish, nil
	default:
		return "", fmt.Errorf("unsupported shell: %s. Supported: bash, zsh, fish", base)
	}
}

func installPath(sh Shell) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home directory: %w", err)
	}

	switch sh {
	case Bash:
		dataDir := os.Getenv("XDG_DATA_HOME")
		if dataDir == "" {
			dataDir = filepath.Join(home, ".local", "share")
		}
		return filepath.Join(dataDir, "bash-completion", "completions", "cwai"), nil
	case Zsh:
		return filepath.Join(home, ".zsh", "completions", "_cwai"), nil
	case Fish:
		configDir := os.Getenv("XDG_CONFIG_HOME")
		if configDir == "" {
			configDir = filepath.Join(home, ".config")
		}
		return filepath.Join(configDir, "fish", "completions", "cwai.fish"), nil
	default:
		return "", fmt.Errorf("unsupported shell: %s", sh)
	}
}

func generateScript(rootCmd *cobra.Command, sh Shell) (string, error) {
	var buf strings.Builder
	var err error

	switch sh {
	case Bash:
		err = rootCmd.GenBashCompletionV2(&buf, true)
	case Zsh:
		err = rootCmd.GenZshCompletion(&buf)
	case Fish:
		err = rootCmd.GenFishCompletion(&buf, true)
	default:
		return "", fmt.Errorf("unsupported shell: %s", sh)
	}

	if err != nil {
		return "", fmt.Errorf("generate completion script: %w", err)
	}

	return marker + "\n" + buf.String(), nil
}

func Install(rootCmd *cobra.Command, shellOverride string) (Shell, string, error) {
	sh, err := DetectShell(shellOverride)
	if err != nil {
		return "", "", err
	}

	path, err := installPath(sh)
	if err != nil {
		return "", "", err
	}

	if data, err := os.ReadFile(path); err == nil {
		if strings.Contains(string(data), marker) {
			return "", "", fmt.Errorf("cwai completion is already installed for %s", sh)
		}
		return "", "", fmt.Errorf("completion file already exists at %s (not managed by cwai). Remove it first", path)
	}

	script, err := generateScript(rootCmd, sh)
	if err != nil {
		return "", "", err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "", fmt.Errorf("create directory %s: %w", dir, err)
	}

	if err := os.WriteFile(path, []byte(script), 0o644); err != nil {
		return "", "", fmt.Errorf("write completion file: %w", err)
	}

	return sh, path, nil
}

func Uninstall(shellOverride string) (Shell, string, error) {
	sh, err := DetectShell(shellOverride)
	if err != nil {
		return "", "", err
	}

	path, err := installPath(sh)
	if err != nil {
		return "", "", err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", "", fmt.Errorf("no cwai completion found for %s", sh)
		}
		return "", "", fmt.Errorf("read completion file: %w", err)
	}

	if !strings.Contains(string(data), marker) {
		return "", "", fmt.Errorf("completion file at %s was not installed by cwai", path)
	}

	if err := os.Remove(path); err != nil {
		return "", "", fmt.Errorf("remove completion file: %w", err)
	}

	return sh, path, nil
}

func PostInstallHint(sh Shell) string {
	switch sh {
	case Bash:
		return "Completions will be loaded automatically in new shell sessions.\nRestart your shell to activate."
	case Zsh:
		return "To activate, add to your ~/.zshrc (before compinit):\n  fpath=(~/.zsh/completions $fpath)\n  autoload -Uz compinit && compinit\nThen restart your shell or run: source ~/.zshrc"
	case Fish:
		return "Fish completions are loaded automatically. Restart your shell to activate."
	default:
		return ""
	}
}
