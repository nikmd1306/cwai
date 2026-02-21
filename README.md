<p align="center">
  <img src="assets/logo.png" width="128" alt="cwai logo">
</p>

<h1 align="center">cwai</h1>

<p align="center">
  AI-powered conventional commit message generator from staged git changes.
</p>

<p align="center">
  <a href="https://github.com/nikmd1306/cwai/actions/workflows/tests.yml"><img src="https://github.com/nikmd1306/cwai/actions/workflows/tests.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/nikmd1306/cwai/releases/latest"><img src="https://img.shields.io/github/v/release/nikmd1306/cwai" alt="Release"></a>
  <a href="https://pkg.go.dev/github.com/nikmd1306/cwai"><img src="https://pkg.go.dev/badge/github.com/nikmd1306/cwai.svg" alt="Go Reference"></a>
  <a href="https://goreportcard.com/report/github.com/nikmd1306/cwai"><img src="https://goreportcard.com/badge/github.com/nikmd1306/cwai" alt="Go Report Card"></a>
  <a href="LICENSE"><img src="https://img.shields.io/github/license/nikmd1306/cwai" alt="License"></a>
</p>

---

## Features

- Generates [Conventional Commits](https://www.conventionalcommits.org/) messages from `git diff`
- Works with any OpenAI-compatible API (OpenAI, Anthropic, local models)
- Two modes: interactive standalone or silent `prepare-commit-msg` hook
- Structured output support for consistent formatting
- Smart diff truncation to fit model token limits

## Installation

### go install

```bash
go install github.com/nikmd1306/cwai@latest
```

### Binary releases

Download pre-built binaries from [Releases](https://github.com/nikmd1306/cwai/releases).

### From source

```bash
git clone https://github.com/nikmd1306/cwai.git
cd cwai
make install
```

## Quick Start

```bash
# Initial setup (API key, model, language)
cwai setup

# Stage changes and generate commit
git add .
cwai

# Or install as git hook for automatic messages
cwai hook set
```

### Standalone mode

```console
$ cwai
Staged files (3):
  internal/ai/client.go
  internal/prompt/prompt.go
  cmd/root.go

Generated commit message:

  feat(ai): add structured output support for commit generation

[y]es / [e]dit / [r]egenerate / [n]o:
```

### Hook mode

```bash
cwai hook set    # Install prepare-commit-msg hook
cwai hook unset  # Remove hook
```

## Configuration

Config file: `~/.cwai` (INI format). Use `cwai config set KEY VALUE` or `cwai setup`.

| Key | Description | Default |
|-----|-------------|---------|
| `CWAI_API_KEY` | API key (required) | *not set* |
| `CWAI_API_URL` | API base URL | `https://api.openai.com/v1` |
| `CWAI_MODEL` | Model name | `gpt-4o-mini` |
| `CWAI_LANGUAGE` | Commit message language | `en` |
| `CWAI_MAX_TOKENS_INPUT` | Max input tokens for diff | `4096` |
| `CWAI_MAX_TOKENS_OUTPUT` | Max output tokens | `500` |
| `CWAI_TEMPERATURE` | Sampling temperature | *not set* |
| `CWAI_REASONING_EFFORT` | Reasoning effort (for reasoning models) | *not set* |
| `CWAI_VERBOSITY` | Output verbosity | *not set* |
| `CWAI_STRUCTURED_OUTPUT` | Enable structured output (`true`/`false`) | *not set* |

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and guidelines.

## License

[MIT](LICENSE)
