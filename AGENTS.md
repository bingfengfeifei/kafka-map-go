# Repository Guidelines

## Project Structure & Module Organization
- `cmd/server/`: Go entrypoint and embedded frontend assets (`cmd/server/web/` contains the built React bundle).  
- `internal/`: Main Go packages (controllers, services, repositories, DTOs, middleware, utilities).  
- `pkg/database/`: SQLite initialization helpers.  
- `web/`: React/Vite frontend; `web/src/` holds components, hooks, and utilities.  
- `config/`: Default YAML configuration and runtime templates.  
Keep business logic in `internal/service/`, expose it through controllers, and avoid circular imports between packages.

## Build, Test, and Development Commands
- `make install-deps` — install Go modules and frontend dependencies.  
- `make build` / `make run` — compile or run backend with embedded UI.  
- `npm --prefix web run build` — produce production UI bundle (copied to `cmd/server/web/`).  
- `go test ./...` — run all backend unit tests.  
Use `make dev` for simultaneous backend/frontend development (hot reload via Vite).

## Coding Style & Naming Conventions
- Go code: `gofmt`-formatted, `golangci-lint`-compatible; prefer `camelCase` for locals and `PascalCase` for exported symbols.  
- React code: functional components with hooks, JSX indented by two spaces, filenames in `PascalCase` (e.g., `TopicInfo.jsx`).  
- Config/JSON keys stay snake_case to match existing REST payloads.

## Testing Guidelines
- Backend: native Go `testing` with table-driven tests; name files `*_test.go`.  
- Frontend: no formal framework yet—add Vitest/React Testing Library if writing new tests.  
- Run `go test ./...` before committing; ensure Kafka-dependent logic is covered with mocks.

## Commit & Pull Request Guidelines
- Follow imperative, descriptive commit messages (`Update branding and GitHub link in UI`, `Align API responses and frontend handling`).  
- Squash trivial fixups; reference issue IDs in the summary or body when applicable.  
- Pull requests should include: problem summary, solution notes, screenshots for UI changes, and manual test evidence (commands or curl output).  
- Keep PRs scoped: backend/REST contract changes must note required frontend updates and vice versa.

## Security & Configuration Tips
- Store credentials in `config/config.yaml`; never hardcode secrets.  
- Tokens expire per `cache.token_expiration`; adjust cautiously.  
- For SASL/SSL clusters, verify `securityProtocol` and `saslMechanism` match the broker setup before deployment.
