# Contributing to tensor_api

Thanks for your interest in contributing!

## Getting started

- Open an issue first for large changes to align on design/approach.
- Small fixes and improvements are welcome as pull requests.

## Development setup

### Backend (Go)
- Go 1.24+
- Build: `go build ./...`
- Test: `go test ./...`

### Web app (`app/`)
- Node.js 20+ and pnpm
- Setup:
  - `corepack enable`
  - `cd app && pnpm install && pnpm build`

## Commit and PR guidelines

- Prefer clear, focused commits. Conventional Commits format is encouraged.
- Ensure CI passes (Go build/test, app build).
- Update docs when behavior or configuration changes.
- For UI changes, attach screenshots when possible.
- Link related issues using `Fixes #<id>` in your PR description.

## Code of Conduct

Please follow the Code of Conduct. Violations can be reported to `admin@shirosora.cn`. Thank you for helping keep the community welcoming.
