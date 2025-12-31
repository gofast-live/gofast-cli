# GoFast CLI (`gof`) - Context Document

## Overview

The `gof` CLI is a code generation tool that builds Go applications like Lego bricks. It generates a full-stack application with:
- Go backend with ConnectRPC transport
- PostgreSQL database with SQLC
- OAuth authentication
- Optional Svelte frontend client

The CLI uses a **skeleton-based code generation** approach - it copies template files and performs smart token replacements and dynamic content generation.

## Source of Truth: `../gofast-app`

**IMPORTANT:** The `gofast-app` repository (located at `../gofast-app` relative to this CLI) is the **single source of truth** for all templates and integrations. When investigating issues or understanding how generated code should look:

1. **Check `../gofast-app` first** - it contains the complete application with all integrations enabled
2. **Template files live there** - domain services, transport handlers, configs, migrations, etc.
3. **Integration markers** (`GF_STRIPE_START/END`, `GF_EMAIL_START/END`, `GF_FILE_START/END`) wrap optional code
4. **Use `TEST=true`** when running CLI commands locally - this copies from `../gofast-app` instead of downloading

```bash
# Local development - uses ../gofast-app as source
TEST=true go run ./cmd/gof/... init demo
TEST=true go run ../cmd/gof/... add stripe   # from inside demo/
```

## Project Structure

```
cmd/gof/
├── main.go              # Entry point, calls cmd.Execute()
├── cmd/                 # Cobra commands
│   ├── root.go          # Root command
│   ├── init.go          # Project initialization
│   ├── add.go           # Add integrations (stripe, r2, postmark)
│   ├── model.go         # Model generation orchestration
│   ├── model_db.go      # Proto, schema, SQL query generation
│   ├── model_service.go # Domain service layer generation
│   ├── model_test_gen.go # Test fixture generation
│   ├── model_transport.go # ConnectRPC transport generation
│   ├── client.go        # Client service setup
│   ├── auth.go          # Auth command
│   ├── infra.go         # Infrastructure files
│   └── version.go       # Version display
├── integrations/        # Optional integration handlers
│   ├── integrations.go  # Shared helpers (strip, copy, migrate)
│   ├── stripe.go        # Stripe payment integration
│   ├── r2.go            # Cloudflare R2 file storage
│   └── postmark.go      # Postmark email integration
├── config/config.go     # gofast.json configuration management
├── repo/repo.go         # Template repository download
├── svelte/svelte.go     # Svelte client page generation
├── auth/                # Authentication TUI (Bubble Tea)
│   ├── auth.go
│   ├── bubble.go
│   └── config.go
└── build.sh             # Build script
```

## CLI Commands

### `gof init [project_name]`
Sets up a new Go project with:
- Docker Compose (PostgreSQL, services)
- OAuth authentication
- Base project structure

**Prerequisites:** buf, goose, sqlc, docker, docker-compose

**Creates:** `gofast.json` config file with project metadata

### `gof model [name] [columns...]`
Generates a complete CRUD model with all layers.

**Syntax:** `gof model note title:string views:number published_at:date is_active:bool`

**Model name rules:**
- Must be lowercase letters and underscores only (e.g., `user_profile`, `event_log`)
- Must be singular - plural names are rejected with a suggestion (e.g., `trucks` → use `truck`)
- Underscores are converted to CamelCase for Go types (e.g., `event_log` → `EventLog`)

**Column name rules:**
- Must be lowercase letters, numbers, and underscores (e.g., `view_count`, `is_active2`)
- Reserved names rejected: `id`, `user_id`, `created`, `updated` (auto-generated)
- Go keywords rejected: `type`, `func`, `var`, `package`, `map`, `chan`, etc.

**Column types:**
| Type | SQL | Proto | Go | Validation |
|------|-----|-------|-----|------------|
| string | text | string | string | required, minlength(3) |
| number | numeric | string | string | must parse, >= 1 |
| date | timestamptz | string | time.Time | valid date format |
| bool | boolean | bool | bool | none |

**Generates:**
1. **Database layer:** Proto definitions, SQL migrations, SQLC queries
2. **Service layer:** Domain service with validation logic
3. **Transport layer:** ConnectRPC handlers
4. **Tests:** Service tests, transport tests, validation tests
5. **Client (if Svelte):** List page, detail/create page, navigation entry

### `gof client svelte`
Adds a Svelte frontend client service:
- Copies template from downloaded repo
- Updates docker-compose
- Generates CRUD pages for all existing models

### `gof infra`
Adds infrastructure/deployment files:
- `docker-compose.otel.yml` - OpenTelemetry setup
- `infra/` folder - Deployment configurations
- `otel/` folder - Observability configs
- Updates `start.sh` to include infrastructure services
- Marks project as `infraPopulated: true` in config

### `gof add <integration>`
Adds optional integrations to the project. Available integrations:

**`gof add stripe`** - Stripe payment integration:
- Payment domain service (checkout, portal, webhook handling)
- Subscriptions database migration
- Full subscription-based access control

**`gof add r2`** - Cloudflare R2 file storage:
- File domain service (upload, download, delete via S3 API)
- Files database migration
- File management UI (if client enabled)

**`gof add postmark`** - Postmark email integration:
- Email domain service (send emails with attachments)
- Emails database migration
- Email management UI (if client enabled)

**See:** [Integration Marker System](#integration-marker-system) for how this works.

## Integration Marker System

The CLI uses a **marker-based integration system** for optional features like Stripe payments. This allows clean addition/removal of integrations without hardcoded string replacements.

### How It Works

1. **Reference repo (`gofast-app`) contains ALL integrations** with code wrapped in markers:
   ```go
   // GF_STRIPE_START
   // ... stripe-specific code ...
   // GF_STRIPE_END
   ```

2. **`gofast.json` tracks enabled integrations:**
   ```json
   {
     "projectName": "myapp",
     "integrations": ["stripe"],
     ...
   }
   ```

3. **On `gof init`:**
   - Copy full template from `gofast-app`
   - Read `integrations` from config (empty by default)
   - Strip ALL marker blocks for integrations NOT in the list
   - Result: clean project without optional features

4. **On `gof add <integration>`:**
   - Add integration name to `gofast.json`
   - Copy relevant files from template (domain/, transport/, migrations)
   - Copy files that have markers, strip only OTHER integrations' markers
   - Result: integration code is present with its markers intact

### Marker Naming Convention

Each integration has its own marker prefix (singular form):
- Stripe: `GF_STRIPE_START` / `GF_STRIPE_END`
- Files: `GF_FILE_START` / `GF_FILE_END`
- Email: `GF_EMAIL_START` / `GF_EMAIL_END`

### Files with Integration Markers

Markers exist in these locations:
- `app/service-core/main.go` - imports, deps, route mounting
- `app/service-core/config/config.go` - integration-specific config fields
- `app/service-core/storage/query.sql` - integration queries
- `app/service-core/storage/migrations/` - integration tables
- `proto/v1/main.proto` - service definitions

For Stripe specifically:
- `app/service-core/domain/login/service.go` - `CheckUserAccess()` function

### Benefits

- **Single source of truth:** All code lives in `gofast-app` with markers
- **Config-driven:** `gofast.json` determines what's included
- **Composable:** Multiple integrations can coexist (stripe + analytics)
- **Maintainable:** No hardcoded string replacements, just marker stripping
- **Future-proof:** Adding new integrations = adding new markers

### Implementation

```go
// Strip integrations not in the enabled list
func StripIntegrations(projectPath string, enabledIntegrations []string) error {
    // Walk all files
    // For each GF_*_START marker, extract integration name
    // If integration not in enabledIntegrations, remove the block
}
```

## How Code Generation Works

### Skeleton-Based Generation

1. **Source templates** live in skeleton directories:
   - `app/service-core/domain/skeleton/` - Service layer
   - `app/service-core/transport/skeleton/` - Transport layer
   - `app/service-client/src/routes/(app)/models/skeletons/` - Svelte pages

2. **Token replacement:**
   - `skeleton` → model name (lowercase)
   - `Skeleton` → Model name (capitalized)
   - `skeletons` → pluralized
   - `Skeletons` → pluralized + capitalized

3. **Dynamic content generation** for tests and validation using markers like:
   ```go
   // GF_FIXTURES_START
   // generated code
   // GF_FIXTURES_END
   ```

### Wiring Injection

The CLI uses **marker-based injection** to wire new models into existing code:

**In `app/service-core/transport/server.go`:**
- `GF_TP_IMPORT_SERVICES_START/END` - Service imports
- `GF_TP_HANDLER_FIELDS_START/END` - Handler struct fields
- `GF_TP_ROUTES_START/END` - Route registration

**In `app/service-core/main.go`:**
- `GF_MAIN_INIT_SERVICES_START/END` - Service instantiation
- `GF_MAIN_HANDLER_ARGS_START/END` - Handler constructor args

**In `app/pkg/auth/auth.go`:**
- `GF_ACCESS_FLAGS_START/END` - Permission flags
- `GF_USER_ACCESS_START/END` - User access bitmask

## Test Generation

Tests are generated for Go only (client-side test generation dropped due to complexity).

### Service Tests (`model_test_gen.go`)
Generates in `app/service-core/domain/{model}/{model}_test.go`:
- `makeQuery{Model}(i, userID)` - Factory with dynamic fields
- `makeInsert{Model}Params(userID)` - Insert params builder
- `makeCreate{Model}Req()` - Proto request fixtures
- Zero/invalid variants for validation testing

### Transport Tests
Generates in `app/service-core/transport/{model}/route_test.go`:
- Request fixtures based on column types
- Validation test cases

### Validation Tests
Generates comprehensive table-driven tests:
- String: required, minlength
- Number: parse validation, >= 1
- Date: format validation
- UUID: for edit operations

## Configuration

**`gofast.json`** - Project configuration file:
```json
{
  "projectName": "myapp",
  "services": [
    {"name": "service-core", "port": 8080},
    {"name": "service-client", "port": 3000}
  ],
  "models": [
    {
      "name": "note",
      "columns": [
        {"name": "title", "type": "string"},
        {"name": "content", "type": "string"}
      ]
    }
  ],
  "integrations": ["stripe"],
  "infraPopulated": false
}
```

**Note:** `integrations` is empty by default. Each `gof add <integration>` command adds to this list.

## Demo Project

The `demo` project contains **output generated by the CLI**. It gets recreated each time the CLI changes to verify what we're producing.

Use it to:
1. See what the CLI currently generates
2. Inspect generated code for debugging
3. Regenerate after CLI changes to verify output

### Regenerating Demo

```bash
rm -rf demo
TEST=true go run ./cmd/gof/... init demo
```

The `TEST=true` env var makes the CLI copy from local `../gofast-app` instead of downloading from the network.

### Adding a Model

From inside the demo directory:
```bash
cd demo
TEST=true go run ../cmd/gof/... model note title:string content:string views:number published:date active:bool
```

### Testing Generated Code

**Important:** Tests require PostgreSQL to be running via Docker Compose.

```bash
# Start PostgreSQL first
cd demo && docker compose up postgres -d

# Apply migrations for new models (init applies base migrations automatically)
goose -dir app/service-core/storage/migrations postgres "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" up

# Run tests
cd demo/app/service-core && go test -race ./...

# Stop PostgreSQL when done
cd demo && docker compose stop
```

### Resetting the Database

If you have stale schema from previous test runs, reset the database:

```bash
cd demo
docker compose down -v   # Remove container AND volume (clears all data)
docker compose up postgres -d
sleep 2  # Wait for postgres to start
goose -dir app/service-core/storage/migrations postgres "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" up
```

## Makefile Commands

Generated projects use Makefile commands instead of shell scripts:

| Command | Description |
|---------|-------------|
| `make start` | Start services with Docker Compose |
| `make startc` | Start with client service |
| `make keys` | Generate public/private keys |
| `make sql` | Regenerate SQLC queries |
| `make gen` | Regenerate proto/buf code |
| `make migrate` | Apply database migrations |

## Design Considerations

**Scalability of generation:**
- `gof model` can have ANY combination of columns
- Examples: 5 dates + 6 bools, all strings, mixed types, etc.
- Generation logic must handle all combinations cleanly

**Test generation complexity:**
- Go tests are being generated (service, transport, validation)
- If test generation becomes too complicated to maintain, **drop it**
- Client-side tests already dropped for this reason
- Keep it simple - if it's too hard to generate reliably, it's not worth it

## Key Files for Development

| Purpose | File |
|---------|------|
| Model orchestration | `cmd/model.go` |
| Test generation | `cmd/model_test_gen.go` |
| Service generation | `cmd/model_service.go` |
| Transport generation | `cmd/model_transport.go` |
| Database/Proto | `cmd/model_db.go` |
| Svelte pages | `svelte/svelte.go` |
| Config management | `config/config.go` |
| Integration helpers | `integrations/integrations.go` |
| Stripe integration | `integrations/stripe.go` |
| R2 integration | `integrations/r2.go` |
| Postmark integration | `integrations/postmark.go` |

## Gotchas

**Proto → TypeScript field naming:**
- Proto uses snake_case: `published_at`
- Generated TypeScript uses camelCase: `publishedAt`
- Svelte generation uses `toCamelCase()` for proto field access

**Use config checks, not file existence:**
- Use `config.IsSvelte()` not `os.Stat("app/service-client")`
- Use `config.HasIntegration("stripe")` to check integrations
- Config is the source of truth for what's enabled

**Migration numbering:**
- Always calculate next number dynamically from existing migrations
- Never copy migrations with hardcoded numbers

## Dependencies

- **Cobra:** CLI framework
- **Bubble Tea:** TUI for authentication
- **go-pluralize:** Pluralization of model names
