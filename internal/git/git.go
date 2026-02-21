package git

import (
	"fmt"
	"os/exec"
	"strings"
)

func run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}

func IsRepo() bool {
	_, err := run("rev-parse", "--is-inside-work-tree")
	return err == nil
}

func HasCommits() bool {
	_, err := run("rev-parse", "HEAD")
	return err == nil
}

func StagedFiles() ([]string, error) {
	out, err := run("diff", "--cached", "--name-only")
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}
	return strings.Split(out, "\n"), nil
}

func StagedDiff() (string, error) {
	return run("diff", "--cached")
}

func StagedDiffStat() (string, error) {
	return run("diff", "--cached", "--stat")
}

func StagedDiffForFile(file string) (string, error) {
	return run("diff", "--cached", "--", file)
}

func StagedNumStat() ([]FileStat, error) {
	out, err := run("diff", "--cached", "--numstat")
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}

	var stats []FileStat
	for _, line := range strings.Split(out, "\n") {
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}
		stats = append(stats, FileStat{
			Added:   parts[0],
			Removed: parts[1],
			File:    parts[2],
		})
	}
	return stats, nil
}

type FileStat struct {
	Added   string
	Removed string
	File    string
}

func Commit(message string) error {
	_, err := run("commit", "-m", message)
	return err
}

func HooksPath() (string, error) {
	out, err := run("rev-parse", "--git-path", "hooks")
	if err != nil {
		return "", err
	}
	return out, nil
}

func GetLatestTag() (string, error) {
	return run("describe", "--tags", "--abbrev=0")
}

func GetPreviousTag(tag string) (string, error) {
	return run("describe", "--tags", "--abbrev=0", tag+"^")
}

func GetCommitsBetween(from, to string) ([]string, error) {
	out, err := run("log", from+".."+to, "--format=%s")
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}
	return strings.Split(out, "\n"), nil
}
