package prompt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildChangelogMessages(t *testing.T) {
	commits := "feat(api): add user endpoint\nfix(auth): fix token validation"
	msgs := BuildChangelogMessages("English", commits)

	require.Len(t, msgs, 2)
	assert.Equal(t, "system", msgs[0].Role)
	assert.Equal(t, "user", msgs[1].Role)
	assert.Equal(t, commits, msgs[1].Content)
	assert.Contains(t, msgs[0].Content, "English")
	assert.Contains(t, msgs[0].Content, "release notes")
}

func TestBuildChangelogMessages_Language(t *testing.T) {
	msgs := BuildChangelogMessages("Russian", "chore: bump deps")
	assert.Contains(t, msgs[0].Content, "Russian")
}

func TestGroupCommitsByType(t *testing.T) {
	commits := []string{
		"feat(api): add user endpoint",
		"feat(ui): add dashboard",
		"fix(auth): fix token validation",
		"refactor(db): simplify queries",
		"perf(cache): optimize lookups",
		"chore(deps): bump dependencies",
		"docs(readme): update installation",
	}

	result := GroupCommitsByType(commits)

	assert.Contains(t, result, "## Features")
	assert.Contains(t, result, "## Bug Fixes")
	assert.Contains(t, result, "## Improvements")
	assert.Contains(t, result, "## Other")
	assert.Contains(t, result, "feat(api): add user endpoint")
	assert.Contains(t, result, "fix(auth): fix token validation")
	assert.Contains(t, result, "refactor(db): simplify queries")
	assert.Contains(t, result, "perf(cache): optimize lookups")
	assert.Contains(t, result, "chore(deps): bump dependencies")
}

func TestGroupCommitsByType_EmptySections(t *testing.T) {
	commits := []string{
		"feat(api): add endpoint",
	}

	result := GroupCommitsByType(commits)

	assert.Contains(t, result, "## Features")
	assert.NotContains(t, result, "## Bug Fixes")
	assert.NotContains(t, result, "## Improvements")
	assert.NotContains(t, result, "## Other")
}
