package ai

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode"
)

type CommitMessageResponse struct {
	ChangesSummary        string   `json:"changes_summary"`
	IntroducesNewBehavior bool     `json:"introduces_new_behavior"`
	FixesBrokenBehavior   bool     `json:"fixes_broken_behavior"`
	RestructuresOnly      bool     `json:"restructures_only"`
	TypeReasoning         string   `json:"type_reasoning"`
	Type                  string   `json:"type"`
	Scope                 string   `json:"scope"`
	Description           string   `json:"description"`
	BulletPoints          []string `json:"bullet_points"`
}

func BuildResponseFormat() map[string]any {
	return map[string]any{
		"type": "json_schema",
		"json_schema": map[string]any{
			"name":   "commit_message",
			"strict": true,
			"schema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"changes_summary": map[string]any{
						"type":        "string",
						"description": "List the main changes in 1-2 sentences.",
					},
					"introduces_new_behavior": map[string]any{
						"type":        "boolean",
						"description": "True ONLY if the diff adds behavior the SYSTEM could not do before. Renaming, moving code to new files/structs, or wrapping existing logic in a new type does NOT count.",
					},
					"fixes_broken_behavior": map[string]any{
						"type":        "boolean",
						"description": "True if the diff corrects behavior that was previously wrong, broken, or producing incorrect results.",
					},
					"restructures_only": map[string]any{
						"type":        "boolean",
						"description": "True if the diff reorganizes, renames, or moves existing code without adding new system capabilities. New files/types that wrap existing logic count as restructuring.",
					},
					"type_reasoning": map[string]any{
						"type":        "string",
						"description": "Apply rules: introduces_new_behavior=true -> feat. fixes_broken_behavior=true -> fix. restructures_only=true -> refactor. Explain which rule applied.",
					},
					"type": map[string]any{
						"type":        "string",
						"enum":        []string{"feat", "fix", "refactor", "docs", "style", "test", "chore", "perf", "ci", "build"},
						"description": "MUST follow from boolean guards. If introduces_new_behavior=true, MUST be feat. refactor ONLY if restructures_only=true.",
					},
					"scope": map[string]any{
						"type":        "string",
						"description": "Exactly ONE word, no spaces. MUST NOT repeat the commit type. Good: auth, api, db, ui. Bad: 'auth system', 'tests' when type is test, 'docs' when type is docs.",
					},
					"description": map[string]any{
						"type":        "string",
						"description": "Imperative verb phrase, lowercase, no trailing period, max 50 chars.",
					},
					"bullet_points": map[string]any{
						"type":        []string{"array", "null"},
						"items":       map[string]any{"type": "string"},
						"description": "When 3+ distinct changes, list concisely. Otherwise null. Start each with a verb.",
					},
				},
				"required": []string{
					"changes_summary",
					"introduces_new_behavior",
					"fixes_broken_behavior",
					"restructures_only",
					"type_reasoning",
					"type",
					"scope",
					"description",
					"bullet_points",
				},
				"additionalProperties": false,
			},
		},
	}
}

func ParseCommitMessageJSON(raw string) (CommitMessageResponse, error) {
	var resp CommitMessageResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return resp, fmt.Errorf("parse commit message JSON: %w", err)
	}
	return resp, nil
}

func AssembleCommitMessage(resp CommitMessageResponse) string {
	desc := resp.Description
	if len(desc) > 0 {
		runes := []rune(desc)
		runes[0] = unicode.ToLower(runes[0])
		desc = string(runes)
	}
	desc = strings.TrimRight(desc, ".")
	if len(desc) > 72 {
		desc = desc[:72]
	}

	header := fmt.Sprintf("%s(%s): %s", resp.Type, resp.Scope, desc)

	if len(resp.BulletPoints) == 0 {
		return header
	}

	var sb strings.Builder
	sb.WriteString(header)
	sb.WriteString("\n\n")
	for i, bp := range resp.BulletPoints {
		sb.WriteString("- ")
		sb.WriteString(bp)
		if i < len(resp.BulletPoints)-1 {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

func SupportsStructuredOutput(apiURL, configOverride string) bool {
	switch strings.ToLower(configOverride) {
	case "on":
		return true
	case "off":
		return false
	}

	u := strings.ToLower(apiURL)
	return strings.Contains(u, "openai.com") || strings.Contains(u, "openrouter.ai")
}
