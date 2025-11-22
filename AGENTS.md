# Repository Guidelines

## Project Structure & Module Organization
The monorepo splits Go microservices under `backend/cmd/<service>` (gateway, user, chat, relay, agent, rag, file, plugin, billing, worker). Shared logic stays in `backend/internal` and reusable helpers in `backend/pkg`, while SQL migrations live in `backend/migrations` and service-specific fixtures or integration tests sit in `backend/tests`. The Next.js UI resides in `frontend/src` with assets in `frontend/public` and global config in `frontend/tailwind.config.ts`. Deployment assets live in `deploy/` (Compose stacks, Kubernetes manifests) and operational scripts at the repository root.

## Build, Test, and Development Commands
Key backend commands (run inside `backend/`):
- `make build` – compile all Go services into `backend/bin/`.
- `make run-gateway` or `cd cmd/<service> && go run main.go` – run an individual service.
- `make dev` – boot the local docker-compose stack from `deploy/docker-compose.yml`.
- `make migrate-up` / `make migrate-down` – apply or roll back database migrations via `DATABASE_URL`.

Frontend (`frontend/`):
- `npm run dev` – start the Next.js dev server on port 3000.
- `npm run build && npm run start` – build and serve production assets.
- `npm run lint`, `npm run format`, `npm run type-check` – ensure linting, formatting, and typing.

## Coding Style & Naming Conventions
Go uses gofmt tabs and idiomatic naming (ExportedType, repo-scoped interfaces). Group transport DTOs inside `pkg/api` and keep package boundaries limited via `internal/`. SQL migration filenames follow timestamped snake_case (e.g., `202411051030_add_usage_table.sql`), and env vars are prefixed `OBLIVIOUS_`. Frontend code is TypeScript with 2-space indentation, React components in PascalCase, hooks/components co-located under feature folders, and Tailwind utility classes for styling. Enforce linting with `golangci-lint run ./...` and `npm run lint`; clean formatting through `make fmt` and `npm run format`.

## Testing Guidelines
Backend tests run through `make test` (`go test -v -cover ./...`). Favor table-driven cases, mock external adapters, and keep package coverage near 70% before merging; inspect regressions via `make test-coverage` and open `backend/coverage.html`. Frontend tests use Jest plus Testing Library (`npm run test` or `npm run test:watch`); colocate `*.test.ts(x)` alongside components, prefer behavior-driven names ("renders billing summary when balance loads"), and snapshot only stable UI fragments.

## Commit & Pull Request Guidelines
Commits follow Conventional Commits already present (`refactor: 优化内存释放逻辑`, `chore(deps): …`). Include scopes tied to directories (`feat(gateway):`, `fix(frontend/chat):`) and ensure code is formatted and linted. Pull requests must describe the change, list affected services/modules, reference issues, and attach test or screenshot evidence. Request reviewers for each touched service, confirm CI is green, and add deployment notes when migrations or infra manifests change.

## Configuration & Security Notes
Seed env files from `backend/env.example` or `backend/env.test`, keeping real secrets outside version control. Update `DATABASE_URL`, Redis, and MinIO endpoints per developer machine; never hardcode provider tokens in config or logs. Kubernetes secrets within `deploy/` should stay templated—store actual keys in your secret manager. For new relay adapters, enforce rate limits and redact upstream credentials in request logs before opening a PR.
