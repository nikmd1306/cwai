//go:build eval

package eval

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func LoadSummary(path string) (RunSummary, error) {
	var s RunSummary
	data, err := os.ReadFile(path)
	if err != nil {
		return s, fmt.Errorf("read summary: %w", err)
	}
	if err := json.Unmarshal(data, &s); err != nil {
		return s, fmt.Errorf("parse summary: %w", err)
	}
	return s, nil
}

func CompareRuns(pathA, pathB string) (string, error) {
	a, err := LoadSummary(filepath.Join(pathA, "summary.json"))
	if err != nil {
		return "", fmt.Errorf("load run A: %w", err)
	}
	b, err := LoadSummary(filepath.Join(pathB, "summary.json"))
	if err != nil {
		return "", fmt.Errorf("load run B: %w", err)
	}
	return FormatComparison(a, b), nil
}

func FormatSummary(s RunSummary) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Run: %s\n", s.RunID)
	fmt.Fprintf(&sb, "Model: %s\n", s.Model)
	fmt.Fprintf(&sb, "Samples: %d (errors: %d)\n", s.TotalSamples, s.ErrorCount)
	fmt.Fprintf(&sb, "Type accuracy:       %.1f%%\n", s.TypeAccuracy*100)
	fmt.Fprintf(&sb, "Scope accuracy:      %.1f%%\n", s.ScopeAccuracy*100)
	fmt.Fprintf(&sb, "Format compliance:   %.1f%%\n", s.FormatCompliance*100)
	fmt.Fprintf(&sb, "Avg desc similarity: %.2f\n", s.AvgDescSimilarity)
	fmt.Fprintf(&sb, "Avg latency (ms):    %.0f\n", s.AvgLatencyMs)
	fmt.Fprintf(&sb, "Total tokens:        %d\n", s.TotalTokens)
	return sb.String()
}

func FormatComparison(a, b RunSummary) string {
	var sb strings.Builder

	labelA := "Run A (" + a.RunID + ")"
	labelB := "Run B (" + b.RunID + ")"

	fmt.Fprintf(&sb, "%-22s %-22s %-22s %s\n", "Metric", labelA, labelB, "Delta")
	sb.WriteString(strings.Repeat("-", 80))
	sb.WriteString("\n")

	pctRow := func(name string, va, vb float64) {
		delta := vb - va
		sign := "+"
		if delta < 0 {
			sign = ""
		}
		fmt.Fprintf(&sb, "%-22s %-22s %-22s %s%.1f%%\n",
			name,
			fmt.Sprintf("%.1f%%", va),
			fmt.Sprintf("%.1f%%", vb),
			sign, delta)
	}

	floatRow := func(name string, va, vb float64) {
		delta := vb - va
		sign := "+"
		if delta < 0 {
			sign = ""
		}
		fmt.Fprintf(&sb, "%-22s %-22s %-22s %s%.2f\n",
			name,
			fmt.Sprintf("%.2f", va),
			fmt.Sprintf("%.2f", vb),
			sign, delta)
	}

	intRow := func(name string, va, vb int) {
		delta := vb - va
		sign := "+"
		if delta < 0 {
			sign = ""
		}
		fmt.Fprintf(&sb, "%-22s %-22s %-22s %s%d\n",
			name,
			fmt.Sprintf("%d", va),
			fmt.Sprintf("%d", vb),
			sign, delta)
	}

	pctRow("Type accuracy", a.TypeAccuracy*100, b.TypeAccuracy*100)
	pctRow("Scope accuracy", a.ScopeAccuracy*100, b.ScopeAccuracy*100)
	pctRow("Format compliance", a.FormatCompliance*100, b.FormatCompliance*100)
	floatRow("Avg desc similarity", a.AvgDescSimilarity, b.AvgDescSimilarity)
	intRow("Avg latency (ms)", int(a.AvgLatencyMs), int(b.AvgLatencyMs))
	intRow("Total tokens", a.TotalTokens, b.TotalTokens)
	intRow("Errors", a.ErrorCount, b.ErrorCount)

	return sb.String()
}
