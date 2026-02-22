---
description: Squash merge a pull request with conventional commit message
argument-hint: <pr-number>
allowed-tools: Bash(git status:*), Bash(git log:*), Bash(git diff:*), Bash(git branch:*), Bash(gh pr view:*), Bash(gh pr list:*), Bash(gh pr merge:*), Bash(gh repo view:*), Read, Grep, Glob
---

## Context

- Repository: !`gh repo view --json nameWithOwner -q .nameWithOwner`
- PR number: $ARGUMENTS

## Task

Perform a squash merge of PR #$ARGUMENTS into `main` with a well-crafted conventional commit message.

### Step 1: Fetch PR Information

Get comprehensive PR data:

```bash
gh pr view $ARGUMENTS --json title,body,commits,headRefName,state,mergeable
```

### Step 2: Analyze All Commits

Review ALL commits in the PR, not just the PR description. The PR description may be outdated if significant changes were made after creation.

For each commit, note:
- The type of change (feat, fix, refactor, etc.)
- The affected scope/package
- Key changes made

### Step 3: Determine Commit Type

Based on the overall changes, select the primary type:

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

If the PR contains multiple types, use the most significant one (typically `feat` > `fix` > `refactor` > others).

### Step 4: Determine Scope

Identify the primary area of changes from the commits:
- `ai` - AI client, schema, structured output
- `config` - Configuration loading/saving
- `git` - Git command wrappers
- `diff` - Diff processing and token estimation
- `prompt` - Prompt construction and templates
- `hook` - Git hook installation
- `completion` - Shell completion
- `changelog` - Changelog generation
- `update` - Self-update mechanism
- `cmd` - CLI command structure
- `release` - GoReleaser, build, distribution
- (or derive from changed files/packages)

### Step 5: Craft Commit Message

**Title format:** `<type>(<scope>): <short description>`
- Max 72 characters
- Lowercase, no period at end
- No emojis
- Should summarize the overall PR purpose

**Body format:**
- Start with a one-sentence summary
- List key changes as bullet points
- Group by category if multiple areas changed
- Include any breaking changes or important notes

### Step 6: Verify PR State

Before merging, ensure:
1. PR state is "OPEN"
2. PR is mergeable
3. No blocking issues

If there are problems, inform the user and stop.

### Step 7: Execute Squash Merge

Use `gh pr merge` with proper formatting:

```bash
gh pr merge $ARGUMENTS --squash --subject "<title>" --body "$(cat <<'EOF'
<body content>
EOF
)"
```

### Step 8: Confirm Result

After merging:
1. Verify the PR state changed to "MERGED"
2. Report the merge commit SHA
3. Provide a link to the merged PR

## Important

- ALWAYS analyze commits thoroughly - PR descriptions often become stale
- Ask for confirmation before merging if the analysis is uncertain
- If PR has conflicts or is not mergeable, inform the user
- Never force merge or skip checks unless explicitly requested
