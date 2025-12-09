# GoFast CLI (`gof`) - Context Document

## Overview

The `gof` CLI is a code generation tool that builds Go applications like Lego bricks. It generates a full-stack application with:
- Go backend with ConnectRPC transport
- PostgreSQL database with SQLC
- OAuth authentication
- Optional Svelte frontend client

The CLI uses a **skeleton-based code generation** approach - it copies template files and performs smart token replacements and dynamic content generation.

## Project Structure

```
cmd/gof/
├── main.go              # Entry point, calls cmd.Execute()
├── cmd/                 # Cobra commands
│   ├── root.go          # Root command
│   ├── init.go          # Project initialization
│   ├── model.go         # Model generation orchestration
│   ├── model_db.go      # Proto, schema, SQL query generation
│   ├── model_service.go # Domain service layer generation
│   ├── model_test_gen.go # Test fixture generation
│   ├── model_transport.go # ConnectRPC transport generation
│   ├── client.go        # Client service setup
│   ├── auth.go          # Auth command
│   ├── infra.go         # Infrastructure files
│   └── version.go       # Version display
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

**Prerequisites:** buf, atlas, sqlc, docker, docker-compose

**Creates:** `gofast.json` config file with project metadata

### `gof model [name] [columns...]`
Generates a complete CRUD model with all layers.

**Syntax:** `gof model note title:string views:number published_at:date is_active:bool`

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
  "infraPopulated": false
}
```

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

## Current Focus: Go Test Generation

The current work focuses on generating Go tests for all layers:
- Service layer tests
- Transport layer tests
- Validation tests

Client-side test generation has been dropped (too complicated). If Go test generation proves too complex, it may also be reconsidered.

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

## Dependencies

- **Cobra:** CLI framework
- **Bubble Tea:** TUI for authentication
- **go-pluralize:** Pluralization of model names
