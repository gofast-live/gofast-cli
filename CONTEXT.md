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

## TEST FUCKING EVERYTHING :D

Comprehensive testing of the CLI. Always include client - `run_tests.sh` covers Go build/lint/test + client lint/build + e2e tests.

### How to Test

```bash
# 1. Generate scenario (from gofast-cli root)
cd /home/mat/projects/gofast-cli
rm -rf demo
TEST=true go run ./cmd/gof/... init demo
cd demo
# ... add models/integrations/client ...

# 2. Run full test suite (requires secrets - ask user for them)
# Script is at scripts/run_tests.sh, run from demo directory
CONTEXT=gofast-rc \
GITHUB_CLIENT_ID=<ask_user> \
GITHUB_CLIENT_SECRET=<ask_user> \
GOOGLE_CLIENT_ID=<ask_user> \
GOOGLE_CLIENT_SECRET=<ask_user> \
TWILIO_ACCOUNT_SID=<ask_user> \
TWILIO_AUTH_TOKEN=<ask_user> \
TWILIO_SERVICE_SID=<ask_user> \
PAYMENT_PROVIDER=stripe \
STRIPE_API_KEY=<ask_user> \
STRIPE_PRICE_ID_BASIC=<ask_user> \
STRIPE_PRICE_ID_PRO=<ask_user> \
STRIPE_WEBHOOK_SECRET=<ask_user> \
BUCKET_NAME=gofast \
R2_ACCESS_KEY=<ask_user> \
R2_SECRET_KEY=<ask_user> \
R2_ENDPOINT=<ask_user> \
EMAIL_FROM=admin@gofast.live \
POSTMARK_API_KEY=<ask_user> \
bash scripts/run_tests.sh

# 3. Check e2e results
cat e2e/test-results/.last-run.json
# Should show: {"status": "passed", "failedTests": []}
```

### Test Scenarios

Each scenario should be tested fresh (rm -rf demo first). Always add client for full coverage.

**Model type variations:**
```bash
# All strings
TEST=true go run ../cmd/gof/... model article title:string body:string author:string

# All numbers
TEST=true go run ../cmd/gof/... model metric count:number value:number score:number

# All dates
TEST=true go run ../cmd/gof/... model event start:date end:date reminder:date

# All bools
TEST=true go run ../cmd/gof/... model settings dark_mode:bool notifications:bool auto_save:bool

# Mixed (the classic)
TEST=true go run ../cmd/gof/... model post title:string views:number published_at:date is_active:bool

# Single column each type
TEST=true go run ../cmd/gof/... model tag name:string
TEST=true go run ../cmd/gof/... model counter value:number
TEST=true go run ../cmd/gof/... model deadline due:date
TEST=true go run ../cmd/gof/... model toggle enabled:bool

# Snake_case names
TEST=true go run ../cmd/gof/... model user_profile display_name:string bio:string
TEST=true go run ../cmd/gof/... model event_log event_type:string occurred_at:date
```

**Integration combinations:**
```bash
# Individual
TEST=true go run ../cmd/gof/... add stripe
TEST=true go run ../cmd/gof/... add r2
TEST=true go run ../cmd/gof/... add postmark

# All together
TEST=true go run ../cmd/gof/... add stripe
TEST=true go run ../cmd/gof/... add r2
TEST=true go run ../cmd/gof/... add postmark

# Order variations (client before/after integrations)
TEST=true go run ../cmd/gof/... client svelte
TEST=true go run ../cmd/gof/... add postmark   # This was a bug!
```

**Client timing variations:**
```bash
# CLIENT AT START - client first, then models and integrations
rm -rf demo
TEST=true go run ./cmd/gof/... init demo
cd demo
TEST=true go run ../cmd/gof/... client svelte
TEST=true go run ../cmd/gof/... add r2
TEST=true go run ../cmd/gof/... add postmark
TEST=true go run ../cmd/gof/... model note title:string content:string
TEST=true go run ../cmd/gof/... model event start:date end:date
TEST=true go run ../cmd/gof/... add stripe
./run_tests.sh

# CLIENT IN MIDDLE - some stuff, then client, then more stuff
rm -rf demo
TEST=true go run ./cmd/gof/... init demo
cd demo
TEST=true go run ../cmd/gof/... add r2
TEST=true go run ../cmd/gof/... model note title:string content:string
TEST=true go run ../cmd/gof/... client svelte
TEST=true go run ../cmd/gof/... add postmark
TEST=true go run ../cmd/gof/... add stripe
TEST=true go run ../cmd/gof/... model task description:string due:date priority:number done:bool
./run_tests.sh

# CLIENT AT END - all models and integrations, then client last
rm -rf demo
TEST=true go run ./cmd/gof/... init demo
cd demo
TEST=true go run ../cmd/gof/... add r2
TEST=true go run ../cmd/gof/... add postmark
TEST=true go run ../cmd/gof/... add stripe
TEST=true go run ../cmd/gof/... model article title:string body:string
TEST=true go run ../cmd/gof/... model counter value:number
TEST=true go run ../cmd/gof/... model deadline due:date
TEST=true go run ../cmd/gof/... model toggle enabled:bool
TEST=true go run ../cmd/gof/... client svelte
./run_tests.sh
```

### Known Bug Patterns

Things that have broken before:
- [ ] formatDate function included when model has no date columns
- [ ] Missing trailing comma in nav array when adding integrations
- [ ] Icon imported but nav entry not added
- [ ] Proto field names (snake_case vs camelCase in TypeScript)
- [ ] Permission flags not updated for new models

## Infrastructure & Integrations Gap Analysis

**Problem:**
The `gof infra` command copies infrastructure files from the source template, but these files **do not** account for optional integrations (Stripe, R2, Postmark). This affects both the main Terraform-based deployment and the Kubernetes-manifest-based PR preview environment.

**Affected Areas:**

1.  **Terraform Infrastructure (`infra/`)**:
    -   `service-core.tf`: Missing environment variable mappings for integration secrets and vars.
    -   `variables.tf`: Missing variable definitions (secrets + vars).
    -   `secrets.tf`: Missing Kubernetes secret resources.

2.  **GitHub Workflows (`.github/workflows/`)**:
    -   `terraform.yml`: The reusable workflow for `tf apply` is missing `TF_VAR_` environment variable mappings for integration secrets and vars.
    -   `pr-deploy.yml`: The PR preview workflow is missing `export` statements to pass secrets and vars to `envsubst`.

3.  **PR Environment Manifests (`infra/pr-environment/`)**:
    -   `secrets.yaml`: Missing `stringData` entries for integration secrets.
    -   `service-core.yaml`: Missing `env` entries for integration secrets and vars.

**Required Changes:**

The CLI must dynamically inject configuration into all above files based on enabled integrations. This should happen during `gof infra` (checking enabled integrations) and `gof add` (if infra exists).

**1. Terraform Injection (`infra/`)**

*   **Stripe:**
    *   `variables.tf`: Add secrets `STRIPE_API_KEY`, `STRIPE_WEBHOOK_SECRET` and vars `STRIPE_PRICE_ID_BASIC`, `STRIPE_PRICE_ID_PRO`
    *   `secrets.tf`: Add `kubernetes_secret` "stripe-secrets"
    *   `service-core.tf`: Add `env` blocks mapping secrets + vars to container env

*   **R2:**
    *   `variables.tf`: Add secrets `R2_ACCESS_KEY`, `R2_SECRET_KEY` and vars `R2_ENDPOINT`, `BUCKET_NAME`
    *   `secrets.tf`: Add `kubernetes_secret` "r2-secrets"
    *   `service-core.tf`: Add `env` blocks

*   **Postmark:**
    *   `variables.tf`: Add secret `POSTMARK_API_KEY` and var `EMAIL_FROM`
    *   `secrets.tf`: Add `kubernetes_secret` "postmark-secrets"
    *   `service-core.tf`: Add `env` blocks

**2. GitHub Workflow Injection (`.github/workflows/`)**

*   `terraform.yml`: Inject `TF_VAR_` mappings for secrets and vars (use `secrets.*` for secrets, `vars.*` for vars).
*   `pr-deploy.yml`: Inject `export` statements for secrets and vars into the "Deploy Other Secrets" step.

**3. PR Environment Injection (`infra/pr-environment/`)**

*   `secrets.yaml`: Inject `stringData` for secret values only.
*   `service-core.yaml`: Inject `env` vars using secrets and direct values.

**4. User Instructions**

*   `gof infra` and `gof add` should print clear instructions telling the user which secrets and vars to add to their GitHub repository environment.
*   Do not update `infra/setup_gh.sh` or `infra/README.md`; treat those as base setup. The CLI will prompt for integration-specific secrets/vars.
