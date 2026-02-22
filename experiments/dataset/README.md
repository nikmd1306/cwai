# Golden Dataset

Evaluation dataset for CWAI commit message generation quality.

## Schema

Each line in `golden.jsonl` is a JSON object:

```json
{
  "id": "unique-slug",
  "description": "Human-readable description of the test case",
  "diff": "Full git diff text",
  "expected_message": "type(scope): expected commit message",
  "expected_type": "feat|fix|refactor|docs|style|test|chore|perf|ci|build",
  "expected_scope": "scope-word",
  "tags": ["optional", "tags"]
}
```

## Fields

| Field | Required | Description |
|-------|----------|-------------|
| `id` | yes | Unique identifier slug (e.g., `add-rate-limiter`) |
| `description` | yes | What this test case covers |
| `diff` | yes | Git diff text that would be input to CWAI |
| `expected_message` | yes | Ideal commit message in conventional commit format |
| `expected_type` | yes | Expected commit type (`feat`, `fix`, etc.) |
| `expected_scope` | yes | Expected scope word |
| `tags` | no | Tags for filtering (e.g., `["multi-file", "feat"]`) |

## Adding Entries

1. Create a realistic git diff (from real commits or synthetic)
2. Write the ideal conventional commit message independently
3. Validate format: `type(scope): description` with lowercase type, single-word scope
4. Add appropriate tags for filtering
5. Append as a single JSON line to `golden.jsonl`

## Tags Convention

- Commit types: `feat`, `fix`, `refactor`, `docs`, `test`, `chore`, `ci`, `style`, `perf`, `build`
- Complexity: `single-file`, `multi-file`, `large-diff`
- Special: `edge-case`, `initial-commit`, `breaking-change`

## Running Evaluation

```bash
make eval                              # Run full evaluation
make eval-test                         # Run eval unit tests
make eval-compare RUN_A=<path> RUN_B=<path>  # Compare two runs
```
