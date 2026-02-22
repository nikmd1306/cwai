# CLAUDE.md

## Project

CWAI (Commits with AI) — Go CLI tool that generates conventional commit messages from staged git changes using OpenAI-compatible APIs.

## Build & Run

```bash
make build      # Build binary
make install    # Install to $GOPATH/bin
make clean      # Remove compiled binary
go test ./...   # Run all tests
```

## Architecture

Entry point: `main.go` → `cmd.Execute()` (Cobra CLI).

### Command Structure (Cobra, `cmd/`)

| Command | File | Description |
|---------|------|-------------|
| `cwai` | `cmd/root.go` | Main command — generates commit message (interactive or hook mode) |
| `cwai setup` | `cmd/setup.go` | Interactive config wizard |
| `cwai config set/get` | `cmd/config.go` | Read/write config keys |
| `cwai hook set/unset` | `cmd/hook.go` | Install/remove `prepare-commit-msg` git hook |

### Internal Packages (`internal/`)

| Package | Purpose |
|---------|---------|
| `ai/client.go` | OpenAI-compatible API client. Detects model families (standard, reasoning, GPT-5) and adapts token parameters per provider. |
| `ai/schema.go` | Structured output JSON schema (`CommitMessageResponse`), parsing, and commit message assembly. |
| `config/config.go` | Loads/saves INI config from `~/.cwai`. Keys prefixed `CWAI_*`. |
| `git/git.go` | Git command wrappers (staged files, diff, commit, hooks path). |
| `diff/diff.go` | Token estimation (length/4 heuristic) and diff truncation to fit token budget. |
| `prompt/prompt.go` | System prompts with few-shot examples for standard and structured output modes. |
| `hook/hook.go` | `prepare-commit-msg` hook installation with marker-based tracking. |

### Data Flow

```
Staged changes → git diff → token truncation → prompt construction → AI API call → structured output parsing → commit message → user interaction (accept/edit/regenerate/abort)
```

### Two Execution Modes

1. **Standalone** (`cwai`) — interactive: shows staged files, generates message, prompts `[y]es/[e]dit/[r]egenerate/[n]o`
2. **Hook** (`cwai --hook <file> [source]`) — silent: writes message to file, skips merge/squash/commit sources

## Configuration

INI file at `~/.cwai`. Key settings: `CWAI_API_KEY`, `CWAI_API_URL`, `CWAI_MODEL`, `CWAI_LANGUAGE`, `CWAI_MAX_TOKENS_INPUT`, `CWAI_MAX_TOKENS_OUTPUT`, `CWAI_TEMPERATURE`, `CWAI_REASONING_EFFORT`, `CWAI_VERBOSITY`, `CWAI_STRUCTURED_OUTPUT`.

## Commits

**IMPORTANT: NEVER use `git commit` directly. Always use `cwai` to create commits.**

Stage files with `git add`, then run `cwai` to generate and apply the commit message. Use `cwai --yes` (or `cwai -y`) for non-interactive mode (CI, scripts, agents). This applies to all commits without exception — single or multiple, interactive or automated.

## Merge Strategy

Squash and merge only. Every PR merges as a single squashed commit into `main`.

## Conventions

- Conventional Commits format: `type(scope): description` + optional bullet points
- No emojis in commit messages
- Go 1.23.4, dependencies: `cobra`, `ini.v1`, `testify` (tests)
- Multi-platform release via GoReleaser (`.goreleaser.yaml`)
