---
description: Create a pull request with analysis of commits
allowed-tools: Bash(git status:*), Bash(git log:*), Bash(git diff:*), Bash(git branch:*), Bash(git show:*), Bash(git rev-parse:*), Bash(git push:*), Bash(gh pr view:*), Bash(gh pr list:*), Bash(gh repo view:*), Read, Grep, Glob
---

## Context

- Current branch: !`git branch --show-current`
- Branch status: !`git status --short`
- Commits since main: !`git log main..HEAD --oneline`
- Changed files: !`git diff main...HEAD --stat | tail -20`

## Task

Create a GitHub pull request from the current branch to `main`.

### Step 1: Analyze Changes

1. Review the commits listed above
2. Read the actual diffs to understand what changed: `git diff main...HEAD`
3. Identify the primary purpose of the changes

### Step 2: Determine PR Type

Based on commit messages and changes, determine the appropriate type:

| Type | When to use |
|------|-------------|
| `feat` | New feature or capability |
| `fix` | Bug fix |
| `refactor` | Code restructuring without behavior change |
| `chore` | Maintenance, dependencies, config |
| `ci` | CI/CD pipeline changes |
| `docs` | Documentation only |
| `test` | Adding or updating tests only |
| `perf` | Performance improvements |

### Step 3: Determine Scope

Identify the primary area of changes:
- `ai` - AI client, schema, structured output
- `config` - Configuration loading/saving
- `git` - Git command wrappers
- `diff` - Diff processing and token estimation
- `prompt` - Prompt construction and templates
- `hook` - Git hook installation
- `completion` - Shell completion
- `changelog` - Changelog generation
- `update` - Self-update mechanism
- `cmd` - CLI command structure (root, setup, config commands)
- `release` - GoReleaser, build, distribution
- (or derive from changed files/packages)

### Step 4: Generate PR

**Title format:** `<type>(<scope>): <short description>`
- Max 72 characters
- Lowercase, no period at end
- No emojis
- Examples:
  - `feat(completion): add shell completion support`
  - `fix(ai): handle empty responses from reasoning models`
  - `chore(release): update goreleaser config`

**Body format:**
```markdown
## Summary
- Bullet points of key changes (3-5 items)

## Key Features (for feat) / Changes (for refactor) / Fixes (for fix)
- Detailed list of what was added/changed/fixed

## Test plan
- [ ] Checklist of testing steps
- [ ] Reference to test files if added
```

### Step 5: Create PR

Use `gh pr create` with HEREDOC for proper formatting:

```bash
gh pr create --base main --title "<title>" --body "$(cat <<'EOF'
<body content>
EOF
)"
```

### Step 6: Return Result

After creating the PR, return the URL so user can review it.

## Important

- Ask for confirmation before creating PR if anything is unclear
- If branch has uncommitted changes, warn the user first
- If branch is not pushed, push it first with `git push -u origin <branch>`
