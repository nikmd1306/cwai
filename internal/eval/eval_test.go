//go:build eval

package eval

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadDataset(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.jsonl")

	samples := []Sample{
		{ID: "s1", Description: "test 1", Diff: "diff1", ExpectedMessage: "feat(api): add endpoint", ExpectedType: "feat", ExpectedScope: "api"},
		{ID: "s2", Description: "test 2", Diff: "diff2", ExpectedMessage: "fix(auth): fix login", ExpectedType: "fix", ExpectedScope: "auth", Tags: []string{"auth"}},
	}

	var content string
	for _, s := range samples {
		line, _ := json.Marshal(s)
		content += string(line) + "\n"
	}
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

	loaded, err := LoadDataset(path)
	require.NoError(t, err)
	assert.Len(t, loaded, 2)
	assert.Equal(t, "s1", loaded[0].ID)
	assert.Equal(t, "s2", loaded[1].ID)
	assert.Equal(t, "feat", loaded[0].ExpectedType)
	assert.Equal(t, []string{"auth"}, loaded[1].Tags)
}

func TestLoadDatasetEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.jsonl")
	require.NoError(t, os.WriteFile(path, []byte(""), 0o644))

	loaded, err := LoadDataset(path)
	require.NoError(t, err)
	assert.Empty(t, loaded)
}

func TestLoadDatasetInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.jsonl")
	require.NoError(t, os.WriteFile(path, []byte("not json\n"), 0o644))

	_, err := LoadDataset(path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "line 1")
}

func TestLoadDatasetFiltered(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.jsonl")

	samples := []Sample{
		{ID: "s1", Diff: "d1", ExpectedMessage: "feat(api): m1", ExpectedType: "feat", ExpectedScope: "api", Tags: []string{"api"}},
		{ID: "s2", Diff: "d2", ExpectedMessage: "fix(auth): m2", ExpectedType: "fix", ExpectedScope: "auth", Tags: []string{"auth"}},
		{ID: "s3", Diff: "d3", ExpectedMessage: "feat(core): m3", ExpectedType: "feat", ExpectedScope: "core", Tags: []string{"api", "auth"}},
	}

	var content string
	for _, s := range samples {
		line, _ := json.Marshal(s)
		content += string(line) + "\n"
	}
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

	t.Run("no filter returns all", func(t *testing.T) {
		result, err := LoadDatasetFiltered(path, nil)
		require.NoError(t, err)
		assert.Len(t, result, 3)
	})

	t.Run("filter by single tag", func(t *testing.T) {
		result, err := LoadDatasetFiltered(path, []string{"auth"})
		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "s2", result[0].ID)
		assert.Equal(t, "s3", result[1].ID)
	})

	t.Run("filter by multiple tags", func(t *testing.T) {
		result, err := LoadDatasetFiltered(path, []string{"api"})
		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "s1", result[0].ID)
		assert.Equal(t, "s3", result[1].ID)
	})

	t.Run("filter with no matches", func(t *testing.T) {
		result, err := LoadDatasetFiltered(path, []string{"nonexistent"})
		require.NoError(t, err)
		assert.Empty(t, result)
	})
}

func TestValidateFormat(t *testing.T) {
	tests := []struct {
		name  string
		msg   string
		valid bool
	}{
		{"simple feat", "feat(api): add endpoint", true},
		{"fix with scope", "fix(auth): fix login bug", true},
		{"refactor", "refactor(core): restructure modules", true},
		{"chore", "chore(deps): update dependencies", true},
		{"hyphenated scope", "ci(github-actions): add workflow", true},
		{"with body", "feat(api): add endpoint\n\n- detail 1\n- detail 2", true},
		{"no conventional format", "invalid message", false},
		{"capitalized type", "Feat(api): capitalized type", false},
		{"no scope", "feat: no scope", false},
		{"empty scope", "feat(): empty scope", false},
		{"no space after colon", "feat(api):no space", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.valid, ValidateFormat(tt.msg))
		})
	}
}

func TestParseCommitHeader(t *testing.T) {
	t.Run("valid header", func(t *testing.T) {
		typ, scope, desc, err := ParseCommitHeader("feat(api): add new endpoint")
		require.NoError(t, err)
		assert.Equal(t, "feat", typ)
		assert.Equal(t, "api", scope)
		assert.Equal(t, "add new endpoint", desc)
	})

	t.Run("with body", func(t *testing.T) {
		typ, scope, desc, err := ParseCommitHeader("fix(auth): fix login\n\n- detail 1\n- detail 2")
		require.NoError(t, err)
		assert.Equal(t, "fix", typ)
		assert.Equal(t, "auth", scope)
		assert.Equal(t, "fix login", desc)
	})

	t.Run("hyphenated scope", func(t *testing.T) {
		typ, scope, desc, err := ParseCommitHeader("ci(github-actions): update workflow")
		require.NoError(t, err)
		assert.Equal(t, "ci", typ)
		assert.Equal(t, "github-actions", scope)
		assert.Equal(t, "update workflow", desc)
	})

	t.Run("invalid format", func(t *testing.T) {
		_, _, _, err := ParseCommitHeader("invalid message")
		assert.Error(t, err)
	})

	t.Run("empty string", func(t *testing.T) {
		_, _, _, err := ParseCommitHeader("")
		assert.Error(t, err)
	})
}

func TestComputeDescSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		a, b     string
		expected float64
	}{
		{"identical", "same string", "same string", 1.0},
		{"both empty", "", "", 1.0},
		{"exact match short", "abc", "abc", 1.0},
		{"completely different", "abc", "xyz", 0.0},
		{"one char diff", "abc", "ab", 2.0 / 3.0},
		{"case insensitive", "Add Endpoint", "add endpoint", 1.0},
		{"with whitespace", "  hello  ", "hello", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ComputeDescSimilarity(tt.a, tt.b)
			assert.InDelta(t, tt.expected, result, 0.01)
		})
	}
}

func TestLevenshtein(t *testing.T) {
	tests := []struct {
		a, b     string
		expected int
	}{
		{"", "", 0},
		{"abc", "", 3},
		{"", "abc", 3},
		{"abc", "abc", 0},
		{"abc", "abd", 1},
		{"abc", "xyz", 3},
		{"kitten", "sitting", 3},
	}

	for _, tt := range tests {
		t.Run(tt.a+"_vs_"+tt.b, func(t *testing.T) {
			assert.Equal(t, tt.expected, levenshtein(tt.a, tt.b))
		})
	}
}

func TestSummarize(t *testing.T) {
	results := []SampleResult{
		{SampleID: "s1", TypeMatch: true, ScopeMatch: true, FormatValid: true, DescSimilarity: 0.8, LatencyMs: 1000, TokensUsed: 100},
		{SampleID: "s2", TypeMatch: true, ScopeMatch: false, FormatValid: true, DescSimilarity: 0.6, LatencyMs: 1200, TokensUsed: 150},
		{SampleID: "s3", TypeMatch: false, ScopeMatch: true, FormatValid: false, DescSimilarity: 0.4, LatencyMs: 800, TokensUsed: 80},
		{SampleID: "s4", Error: "api error"},
	}

	summary := Summarize("test-run", "gpt-4", "https://api.openai.com/v1", true, results)

	assert.Equal(t, "test-run", summary.RunID)
	assert.Equal(t, "gpt-4", summary.Model)
	assert.Equal(t, true, summary.StructuredOutput)
	assert.Equal(t, 4, summary.TotalSamples)
	assert.Equal(t, 1, summary.ErrorCount)
	assert.InDelta(t, 2.0/3.0, summary.TypeAccuracy, 0.01)
	assert.InDelta(t, 2.0/3.0, summary.ScopeAccuracy, 0.01)
	assert.InDelta(t, 2.0/3.0, summary.FormatCompliance, 0.01)
	assert.InDelta(t, 0.6, summary.AvgDescSimilarity, 0.01)
	assert.InDelta(t, 1000.0, summary.AvgLatencyMs, 0.01)
	assert.Equal(t, 330, summary.TotalTokens)
}

func TestSummarizeEmpty(t *testing.T) {
	summary := Summarize("empty-run", "gpt-4", "https://api.openai.com/v1", false, nil)
	assert.Equal(t, 0, summary.TotalSamples)
	assert.Equal(t, 0.0, summary.TypeAccuracy)
}

func TestSummarizeAllErrors(t *testing.T) {
	results := []SampleResult{
		{SampleID: "s1", Error: "error 1"},
		{SampleID: "s2", Error: "error 2"},
	}
	summary := Summarize("error-run", "gpt-4", "https://api.openai.com/v1", false, results)
	assert.Equal(t, 2, summary.TotalSamples)
	assert.Equal(t, 2, summary.ErrorCount)
	assert.Equal(t, 0.0, summary.TypeAccuracy)
}

func TestFormatSummary(t *testing.T) {
	s := RunSummary{
		RunID:             "test-run",
		Model:             "gpt-4",
		TotalSamples:      10,
		ErrorCount:        1,
		TypeAccuracy:      0.9,
		ScopeAccuracy:     0.8,
		FormatCompliance:  1.0,
		AvgDescSimilarity: 0.75,
		AvgLatencyMs:      1200,
		TotalTokens:       5000,
	}
	output := FormatSummary(s)
	assert.Contains(t, output, "test-run")
	assert.Contains(t, output, "gpt-4")
	assert.Contains(t, output, "90.0%")
	assert.Contains(t, output, "100.0%")
}

func TestFormatComparison(t *testing.T) {
	a := RunSummary{
		RunID:             "run-a",
		TypeAccuracy:      0.85,
		ScopeAccuracy:     0.70,
		FormatCompliance:  1.0,
		AvgDescSimilarity: 0.62,
		AvgLatencyMs:      1200,
		TotalTokens:       5000,
		ErrorCount:        0,
	}
	b := RunSummary{
		RunID:             "run-b",
		TypeAccuracy:      0.92,
		ScopeAccuracy:     0.80,
		FormatCompliance:  1.0,
		AvgDescSimilarity: 0.71,
		AvgLatencyMs:      1350,
		TotalTokens:       5200,
		ErrorCount:        1,
	}
	output := FormatComparison(a, b)
	assert.Contains(t, output, "Type accuracy")
	assert.Contains(t, output, "run-a")
	assert.Contains(t, output, "run-b")
	assert.Contains(t, output, "Delta")
}

func TestLoadSummary(t *testing.T) {
	dir := t.TempDir()
	s := RunSummary{
		RunID:        "test",
		Model:        "gpt-4",
		TotalSamples: 5,
		TypeAccuracy: 0.8,
	}
	data, _ := json.MarshalIndent(s, "", "  ")
	path := filepath.Join(dir, "summary.json")
	require.NoError(t, os.WriteFile(path, data, 0o644))

	loaded, err := LoadSummary(path)
	require.NoError(t, err)
	assert.Equal(t, "test", loaded.RunID)
	assert.Equal(t, 5, loaded.TotalSamples)
	assert.InDelta(t, 0.8, loaded.TypeAccuracy, 0.01)
}

func TestCompareRuns(t *testing.T) {
	dir := t.TempDir()

	runA := filepath.Join(dir, "run-a")
	runB := filepath.Join(dir, "run-b")
	require.NoError(t, os.MkdirAll(runA, 0o755))
	require.NoError(t, os.MkdirAll(runB, 0o755))

	sA := RunSummary{RunID: "run-a", TypeAccuracy: 0.8}
	sB := RunSummary{RunID: "run-b", TypeAccuracy: 0.9}

	dataA, _ := json.MarshalIndent(sA, "", "  ")
	dataB, _ := json.MarshalIndent(sB, "", "  ")

	require.NoError(t, os.WriteFile(filepath.Join(runA, "summary.json"), dataA, 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(runB, "summary.json"), dataB, 0o644))

	result, err := CompareRuns(runA, runB)
	require.NoError(t, err)
	assert.Contains(t, result, "run-a")
	assert.Contains(t, result, "run-b")
	assert.Contains(t, result, "Type accuracy")
}
