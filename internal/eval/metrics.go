//go:build eval

package eval

import (
	"fmt"
	"regexp"
	"strings"
)

type SampleResult struct {
	SampleID         string  `json:"sample_id"`
	GeneratedMessage string  `json:"generated_message"`
	ExpectedMessage  string  `json:"expected_message"`
	TypeMatch        bool    `json:"type_match"`
	ScopeMatch       bool    `json:"scope_match"`
	FormatValid      bool    `json:"format_valid"`
	DescSimilarity   float64 `json:"desc_similarity"`
	TokensUsed       int     `json:"tokens_used"`
	LatencyMs        int64   `json:"latency_ms"`
	Error            string  `json:"error,omitempty"`
}

type RunSummary struct {
	RunID             string  `json:"run_id"`
	Model             string  `json:"model"`
	APIURL            string  `json:"api_url"`
	StructuredOutput  bool    `json:"structured_output"`
	Timestamp         string  `json:"timestamp"`
	TotalSamples      int     `json:"total_samples"`
	TypeAccuracy      float64 `json:"type_accuracy"`
	ScopeAccuracy     float64 `json:"scope_accuracy"`
	FormatCompliance  float64 `json:"format_compliance"`
	AvgDescSimilarity float64 `json:"avg_desc_similarity"`
	AvgLatencyMs      float64 `json:"avg_latency_ms"`
	TotalTokens       int     `json:"total_tokens"`
	ErrorCount        int     `json:"error_count"`
}

var commitHeaderRegex = regexp.MustCompile(`^[a-z]+\([a-z0-9-]+\): .+`)

func ValidateFormat(message string) bool {
	firstLine := strings.SplitN(message, "\n", 2)[0]
	return commitHeaderRegex.MatchString(firstLine)
}

func ParseCommitHeader(message string) (typ, scope, desc string, err error) {
	firstLine := strings.SplitN(message, "\n", 2)[0]
	re := regexp.MustCompile(`^([a-z]+)\(([a-z0-9-]+)\): (.+)$`)
	matches := re.FindStringSubmatch(firstLine)
	if matches == nil {
		return "", "", "", fmt.Errorf("invalid commit header format: %q", firstLine)
	}
	return matches[1], matches[2], matches[3], nil
}

func ComputeDescSimilarity(a, b string) float64 {
	a = strings.ToLower(strings.TrimSpace(a))
	b = strings.ToLower(strings.TrimSpace(b))
	if a == b {
		return 1.0
	}
	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}
	if maxLen == 0 {
		return 1.0
	}
	dist := levenshtein(a, b)
	return 1.0 - float64(dist)/float64(maxLen)
}

func levenshtein(a, b string) int {
	ra := []rune(a)
	rb := []rune(b)
	la := len(ra)
	lb := len(rb)

	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	prev := make([]int, lb+1)
	curr := make([]int, lb+1)

	for j := 0; j <= lb; j++ {
		prev[j] = j
	}

	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			curr[j] = min(prev[j]+1, curr[j-1]+1, prev[j-1]+cost)
		}
		prev, curr = curr, prev
	}
	return prev[lb]
}

func Summarize(runID, model, apiURL string, structured bool, results []SampleResult) RunSummary {
	s := RunSummary{
		RunID:            runID,
		Model:            model,
		APIURL:           apiURL,
		StructuredOutput: structured,
		TotalSamples:     len(results),
	}
	if len(results) == 0 {
		return s
	}

	var typeMatches, scopeMatches, formatValid int
	var totalSimilarity float64
	var totalLatency int64

	for _, r := range results {
		if r.Error != "" {
			s.ErrorCount++
			continue
		}
		if r.TypeMatch {
			typeMatches++
		}
		if r.ScopeMatch {
			scopeMatches++
		}
		if r.FormatValid {
			formatValid++
		}
		totalSimilarity += r.DescSimilarity
		totalLatency += r.LatencyMs
		s.TotalTokens += r.TokensUsed
	}

	evaluated := len(results) - s.ErrorCount
	if evaluated > 0 {
		s.TypeAccuracy = float64(typeMatches) / float64(evaluated)
		s.ScopeAccuracy = float64(scopeMatches) / float64(evaluated)
		s.FormatCompliance = float64(formatValid) / float64(evaluated)
		s.AvgDescSimilarity = totalSimilarity / float64(evaluated)
		s.AvgLatencyMs = float64(totalLatency) / float64(evaluated)
	}
	return s
}
