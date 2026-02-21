# Contributing to cwai

Thanks for your interest in contributing! This document covers everything you need to get started.

## Development Setup

### Prerequisites

- Go 1.23.4+
- Git

### Build & test

```bash
git clone https://github.com/nikmd1306/cwai.git
cd cwai

make build       # Build binary
make install     # Install to $GOPATH/bin
make clean       # Remove compiled binary
go test ./...    # Run all tests
```

## How to Contribute

### Reporting bugs

Open an [issue](https://github.com/nikmd1306/cwai/issues/new?template=bug_report.md) with:
- Steps to reproduce
- Expected vs actual behavior
- Go version, OS, cwai version

### Suggesting features

Open a [feature request](https://github.com/nikmd1306/cwai/issues/new?template=feature_request.md) describing the use case and proposed solution.

### Submitting code

1. Fork the repository
2. Create a feature branch: `git checkout -b feat/my-feature`
3. Make your changes
4. Run tests: `go test ./...`
5. Commit your changes — and yes, we expect [Conventional Commits](https://www.conventionalcommits.org/). You know a tool that can help with that, right? `cwai` ;)
6. Push and open a Pull Request

## Pull Request Guidelines

- Conventional Commits for all commit messages (pro tip: `make install && cwai` — dogfood at its finest)
- Add tests for new functionality
- Ensure all existing tests pass
- Keep PRs focused — one feature or fix per PR

## Code Style

- Format code with `gofmt`
- Run `go vet ./...` before committing
- Follow standard Go conventions and project structure
