# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

OTF is an open source alternative to Terraform Enterprise. It provides SSO, team management, agents, and no per-resource pricing. The system manages Terraform/OpenTofu runs, workspaces, and state files.

**Documentation:** https://docs.otf.ninja/

## Architecture

### Core Components

- **otfd** (`cmd/otfd/main.go`): Main daemon that runs the web server, API, and embedded runner
- **otf-agent** (`cmd/otf-agent/main.go`): External agent that connects to otfd to execute runs for a specific organization

### Service Layer Architecture

The codebase follows a domain-driven design with services organized by entity type in `internal/`:

- **run**: Terraform/OpenTofu run lifecycle (plan, apply, status transitions)
- **workspace**: Workspace management and locking
- **organization**: Organization management and membership
- **team**: Team management and permissions
- **module**: Module registry functionality
- **vcs**: VCS provider integrations (GitHub, GitLab, Forgejo)
- **configversion**: Configuration version uploads and management
- **state**: Terraform state management
- **runner**: Agent/runner system for executing terraform commands
- **engine**: Terraform/OpenTofu engine abstraction
- **ui**: Centralized web UI handlers for all services

Each service typically has:
- `service.go`: Core service logic
- `db.go`: Database access layer (PostgreSQL via pgx)
- `api.go`: TFE-compatible JSON API handlers
- `cli.go`: CLI command support

### UI Package Architecture

Web UI handlers have been refactored into a centralized `internal/ui` package:

- **ui/handlers.go**: Main handlers struct that implements `internal.Handlers`
- **ui/run.go**: Run-related UI handlers (listing, viewing, creating runs)
- **ui/run_view.templ**: Templ templates for run UI components
- **ui/run_cache.go**: Caching helpers for UI data fetching

The UI package imports domain services (run, workspace, user) and provides exported registration functions like `AddRunHandlers()`. This approach:
- Centralizes UI code in one package
- Avoids import cycles by having the ui package import domain services (not vice versa)
- Maintains separation between UI layer and business logic
- Allows domain services to focus on their core responsibilities

The `ui.Handlers` struct is registered in the daemon's handlers list (`internal/daemon/daemon.go`) alongside other services.

### Key Architectural Concepts

**Run Queue System** (see `ALGORITHMS.md`):
- Each workspace maintains a queue of runs (workspace queue)
- Runs move from `pending` → `plan_queued` → `planning` → `planned` → `apply_queued` → `applying` → `applied`
- Global queue prioritizes: (1) applies, (2) normal plans, (3) speculative plans
- Speculative runs (plan-only) bypass workspace queue

**Runner/Agent System**:
- otfd has an embedded runner that executes runs for any organization
- External agents authenticate with org-scoped tokens and only execute runs for that organization
- Agents can execute multiple run phases concurrently

**Database Layer**:
- PostgreSQL database managed via tern migrations (`internal/sql/migrations/`)
- Connection pooling via pgx
- Database access centralized in `*db.go` files per service

**Authentication & Authorization**:
- `internal/authenticator`: Handles OAuth (GitHub, GitLab, OIDC), sessions
- `internal/authz`: RBAC system with roles and permissions
- Context-based authorization checks throughout services

**HTTP Layer** (`internal/http/`):
- `html/`: Web UI using templ templates + Tailwind CSS
- TFE-compatible JSON API for terraform CLI integration
- WebSocket support for real-time log streaming

## Development Commands

### Build & Install
```bash
make build          # Build binaries to _build/
make install        # Install binaries to $GOBIN
```

### Testing
```bash
make test                      # Run all unit tests
go test ./internal/run         # Run tests for specific package
go test -run TestName ./...    # Run specific test

# Integration tests (require docker-compose)
make compose-up                # Start postgres, squid, pubsub services
go test ./internal/integration/... -count 1
make compose-rm                # Clean up services

# TFE API compatibility tests
make go-tfe-tests
```

### Code Quality
```bash
make fmt            # Format code with go fmt
make vet            # Run go vet
make lint           # Run staticcheck linter
make install-linter # Install staticcheck first

# Pre-commit hooks (runs fmt, vet, lint automatically)
make install-pre-commit
```

### Frontend Development
```bash
# Live reload development (runs in parallel):
make live           # Watches templ, tailwind, and static assets

# Or run individually:
make live/templ     # Watch .templ files and reload browser
make live/tailwind  # Watch and rebuild Tailwind CSS
make live/sync_assets  # Watch static assets

# Manual template generation:
make generate-templates  # Run templ generate
go tool templ generate   # Direct templ invocation
```

### Code Generation
```bash
make paths          # Generate path helpers for HTML routes
make actions        # Generate RBAC action strings (uses stringer)
```

### Database
```bash
# Database connection string format:
# postgres:///otf?host=/var/run/postgresql  (Unix socket)
# postgres://user:pass@host:5432/otf        (TCP)

make postgres       # Start postgres via docker-compose
make install-migrator  # Install tern migration tool

# Run migrations manually:
tern migrate -c tern.conf
```

### Debugging
```bash
make debug          # Start delve debugger headless on :4300
make connect        # Connect to running delve session
```

### Docker
```bash
make image          # Build otfd Docker image
make image-agent    # Build otf-agent Docker image
make load           # Load otfd image into kind cluster
make load-agent     # Load agent image into kind cluster
```

## Important Implementation Notes

### Working with Services

When modifying a service:
1. Database changes require a new migration in `internal/sql/migrations/`
2. API changes must maintain TFE compatibility
3. Authorization checks use `authz.Interface` - always check permissions before operations
4. Use structured logging via `logr.Logger`

### Testing Patterns

- Unit tests: Mock dependencies, test service logic in isolation
- Integration tests (`internal/integration/`): Full stack tests with real database
- Browser tests: Use Playwright (`internal/testbrowser/`)
- Fixtures: Test terraform configs in `internal/integration/testdata/`

### Frontend Development

- Templates use templ syntax (`.templ` files generate `_templ.go`)
- Styles use Tailwind CSS (`internal/http/html/static/css/`)
- JavaScript uses choices.js for multi-select dropdowns
- Always run `make generate-templates` after modifying `.templ` files
- Use `make live` for hot reload during development

### Database Patterns

- Use pgx for PostgreSQL access (`github.com/jackc/pgx/v5`)
- Transactions are context-based via `sql.DB`
- Use `sql.Listen` for pub/sub notifications
- Run migrations with tern before tests

### Authentication Flow

1. User authenticates via OAuth (GitHub/GitLab) or OIDC
2. Session stored in cookie
3. Context enriched with `authz.Subject` (user, team, or superuser)
4. Authorization checks via `authz.Interface.Authorize(subject, action, resource)`

### VCS Integration

- VCS providers in `internal/github/`, `internal/gitlab/`, `internal/forgejo/`
- Each implements common interfaces for webhooks, OAuth, repo operations
- Workspace can be connected to VCS repo for automatic runs on push
- VCS events handled by event handlers (e.g., `internal/github/event_handler.go`)

### Run Lifecycle Hooks

Services can register hooks for run lifecycle events:
- `afterCancelHooks`
- `afterForceCancelHooks`
- `afterEnqueuePlanHooks`
- `afterEnqueueApplyHooks`

Used for notifications, VCS status updates, etc.

## Common Pitfalls

1. **Database migrations**: Always create a new migration file, never edit existing ones
2. **Templ generation**: Changes to `.templ` files require running `make generate-templates`
3. **Context**: Most functions require `context.Context` - don't use `context.Background()` in handlers
4. **Authorization**: Always verify permissions before performing operations
5. **Testing**: Integration tests need docker-compose services running (`make compose-up`)
6. **Go tools**: Some tools (templ, staticcheck, stringer) are managed via `go.mod` tool directive

## File Naming Conventions

- `*_test.go`: Unit tests
- `*_templ.go`: Generated from `.templ` files (don't edit directly)
- `db.go`: Database access layer for the service
- `service.go`: Main service implementation
- `api.go`: JSON API handlers (TFE-compatible)
- `cli.go`: CLI command implementations

**Note**: HTML web UI handlers are now centralized in `internal/ui/` rather than in individual service `web.go` files.
