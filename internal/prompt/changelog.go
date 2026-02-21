package prompt

import (
	"fmt"
	"strings"
)

const changelogSystemTemplate = `You are a release notes generator. You receive a list of conventional commit messages and produce well-structured Markdown release notes.

RULES:
1. Start with a brief 1-2 sentence summary of the release.
2. Group changes into sections: Features, Bug Fixes, Improvements, Other Changes.
3. Only include sections that have relevant commits.
4. Each item should be a clear, human-readable description derived from the commit message.
5. Do NOT include raw commit messages verbatim — rephrase them for readability.
6. Use Markdown formatting with ## headers and bullet points.
7. NEVER use emojis.
8. Use %s for the release notes language.
9. Output ONLY the Markdown release notes. No explanations, no wrapping.`

func BuildChangelogMessages(language string, commits string) []Message {
	system := fmt.Sprintf(changelogSystemTemplate, language)
	return []Message{
		{Role: "system", Content: system},
		{Role: "user", Content: commits},
	}
}

func GroupCommitsByType(commits []string) string {
	groups := map[string][]string{
		"Features":    {},
		"Bug Fixes":   {},
		"Improvements": {},
		"Other":       {},
	}

	for _, c := range commits {
		switch {
		case strings.HasPrefix(c, "feat"):
			groups["Features"] = append(groups["Features"], c)
		case strings.HasPrefix(c, "fix"):
			groups["Bug Fixes"] = append(groups["Bug Fixes"], c)
		case strings.HasPrefix(c, "refactor") || strings.HasPrefix(c, "perf"):
			groups["Improvements"] = append(groups["Improvements"], c)
		default:
			groups["Other"] = append(groups["Other"], c)
		}
	}

	var sb strings.Builder
	for _, section := range []string{"Features", "Bug Fixes", "Improvements", "Other"} {
		items := groups[section]
		if len(items) == 0 {
			continue
		}
		sb.WriteString("## " + section + "\n\n")
		for _, item := range items {
			sb.WriteString("- " + item + "\n")
		}
		sb.WriteString("\n")
	}
	return strings.TrimSpace(sb.String())
}
