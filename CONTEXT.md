# GoFast CLI (`gof`) Context

## Metadata
- Domain: `gof` CLI - Go application code generator
- Primary audience: LLM agents working on CLI development
- Last updated: 2026-03-10
- Status: Active
- Stability note: Sections marked `[STABLE]` should change rarely. Sections marked `[VOLATILE]` are expected to change often.

---

## 0. Context Maintenance Protocol (LLM-First) [STABLE]

This file is the primary working context for the `gof` CLI tool.

- LLM agents should treat this as a living document and update it whenever meaningful behavior changes.
- If code and this file diverge, prefer updating this file quickly so future work stays reliable.
- Temporary or branch-specific behavior should be documented here with clear cleanup notes.

### Quick update checklist
- Refresh `Last updated` date
- Review `Current Work` and `Future Work`
- Validate `Critical Invariants`
- Update marker references if any markers renamed or added
- Remove obsolete notes

### Freshness target
- Re-review this file regularly (every 2 weeks) to prevent context drift.

---

## 1. Summary [STABLE]

The `gof` CLI is a code generation tool that builds full-stack Go applications like Lego bricks. It generates a production-ready application with Go backend (ConnectRPC), PostgreSQL (SQLC), OAuth auth, optional Svelte or TanStack frontend, and optional integrations (Stripe, S3, Postmark).

The CLI uses **skeleton-based code generation**: it copies template files from a reference repository and performs token replacements plus marker-based dynamic content injection.

- **Primary entry points:** `cmd/gof/main.go` -> `cmd/gof/cmd/root.go` (Cobra commands)
- **Main responsibilities:** Project scaffolding, CRUD model generation, integration wiring, client scaffolding, infra/monitoring setup
- **Highest-risk areas:** Test generation (`model_test_gen.go`), marker-based code injection, multi-client scaffolding/order-dependent generation, permission bitmask computation

---

## Source of Truth: `../gofast-app` [STABLE]

**CRITICAL:** The `gofast-app` repository (at `../gofast-app` relative to this CLI) is the **single source of truth** for all templates and integrations.

1. **Check `../gofast-app` first** when investigating issues or understanding generated code
2. **Skeleton templates live there** - `domain/skeleton/`, `transport/skeleton/`, `e2e/skeletons.test.ts`
3. **Integration markers** (`GF_STRIPE_START/END`, `GF_FILE_START/END`, `GF_EMAIL_START/END`) wrap optional code
4. **Use `TEST=true`** when running CLI commands locally - copies from `../gofast-app` instead of downloading

```bash
# Local development - uses ../gofast-app as source
TEST=true go run ./cmd/gof/... init demo
TEST=true go run ../cmd/gof/... add stripe   # from inside demo/
```

### Key template locations in `../gofast-app`

| Template | Path |
|----------|------|
| Service skeleton | `app/service-core/domain/skeleton/` (service.go, service_test.go, validation.go, validation_test.go) |
| Transport skeleton | `app/service-core/transport/skeleton/` (route.go, route_test.go) |
| E2E skeleton | `e2e/skeletons.test.ts` |
| Proto skeleton | `proto/v1/skeleton.proto` |
| Svelte skeleton | `app/service-svelte/src/routes/(app)/models/skeletons/` |
| TanStack skeleton | `app/service-tanstack/src/routes/_layout/models/skeletons/` |
| Main wiring | `app/service-core/main.go` (marker injection points) |
| Auth permissions | `app/pkg/auth/auth.go` (permission flags) |

---

## Test Strategy [STABLE]

### ALWAYS TEST EVERYTHING

Testing is the most critical part of this CLI. Every change must be verified by generating a demo project and running its full test suite. Always include a client for full coverage.

### Scope and intent
- The CLI itself has no unit tests - it is tested by **generating projects and running their tests**
- Generated projects include Go unit tests, integration tests, and Playwright e2e tests
- CI runs `golangci-lint` on the CLI code itself (`.github/workflows/lint.yml`)

### LLM default policy
- On every code change, regenerate a demo project and verify Go builds + tests pass
- For marker/injection changes, test multiple model type combinations
- For integration changes, test with client enabled and disabled
- If tests are deferred, document the gap explicitly

### How to test

```bash
# 1. Generate scenario (from gofast-cli root)
cd /home/mat/projects/gofast-cli
rm -rf demo
TEST=true go run ./cmd/gof/... init demo
cd demo

# 1.5. For local testing, switch buf.gen.yaml to local plugins
# Replace remote plugins with local ones before running `make gen`
# Use this shape:
# version: v2
# plugins:
#   - local: protoc-gen-go
#     out: app/gen
#     opt: paths=source_relative
#   - local: protoc-gen-connect-go
#     out: app/gen
#     opt: paths=source_relative
#   - local: protoc-gen-es
#     out: app/service-svelte/src/lib/gen
#     opt: target=ts
#   - local: protoc-gen-es
#     out: app/service-tanstack/src/lib/gen
#     opt: target=ts

# ... add models/integrations/client as needed ...

# 2. Quick verification (no secrets needed)
docker compose up postgres -d
goose -dir app/service-core/storage/migrations postgres \
  "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" up
cd app/service-core && go test -race ./...

# 3. Full test suite (requires secrets - ask user for them)
# Run from demo directory
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

# 4. Check e2e results
cat e2e/test-results/.last-run.json
# Should show: {"status": "passed", "failedTests": []}
```

### E2E validation (after client-side generation)

After every client-side generation (`gof client svelte|tanstack`, `gof model` with client enabled, `gof add` with client enabled), validate with the full e2e suite. This is the most important validation indicator — Go unit tests alone don't catch client-side wiring issues.

```bash
# From demo/ directory, after all gof commands and codegen:
make gen && make sql

# Start services (use starts for svelte, startt for tanstack)
docker compose -f docker-compose.yml -f docker-compose.tanstack.yml up --build -d
# or: docker compose -f docker-compose.yml -f docker-compose.svelte.yml up --build -d

# Wait for all containers to be healthy, then migrate
make migrate

# Run e2e tests
cd e2e
PLAYWRIGHT_BASE_URL=http://localhost:3000 PUBLIC_CORE_URL=http://localhost:4000 npx playwright test
```

**Expected results (no integrations):** ~33 passed, ~28 skipped (integration tests for emails/files/payments).
**Expected results (with integrations):** skipped count decreases as integration tests become runnable.

**Common e2e pitfalls:**
- `docker compose down -v` wipes the DB volume — must re-run `make migrate` before tests
- `routeTree.gen.ts` (TanStack) must match actual route files — the CLI now regenerates it via `@tanstack/router-generator` during TanStack client formatting
- Proto codegen (`make gen`) must run before `make startt` — client imports services from `main_pb.ts`
- `e2e/users.test.ts` in the root template is currently flaky because it clicks the first `Edit` link instead of the row for the generated `adminEmail`

### Resetting the database

```bash
cd demo
docker compose down -v   # Remove container AND volume
docker compose up postgres -d
sleep 2
goose -dir app/service-core/storage/migrations postgres \
  "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" up
```

### Test scenarios

Each scenario should be tested fresh (`rm -rf demo` first). Always add client for full coverage.

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
# Mixed (classic)
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
TEST=true go run ../cmd/gof/... add s3
TEST=true go run ../cmd/gof/... add postmark
# All together (order matters - test different orders)
```

**Client timing variations (critical - order bugs are common):**
```bash
# CLIENT AT START
gof client svelte|tanstack -> add integrations -> add models

# CLIENT IN MIDDLE
add some integrations/models -> gof client svelte|tanstack -> add more

# CLIENT AT END
add all integrations/models -> gof client svelte|tanstack
```

### Known bug patterns
- formatDate function included when model has no date columns
- Missing trailing comma in nav array when adding integrations
- Icon imported but nav entry not added
- Proto field names (snake_case vs camelCase in TypeScript)
- Permission flags not updated for new models

---

## 2. Architecture [STABLE]

### 2.1 Component map

```mermaid
flowchart LR
  subgraph CLI["gof CLI"]
    CMD[Cobra Commands]
    CFG[Config Manager]
    INT[Integrations]
    SVL[Svelte Generator]
    E2E[E2E Generator]
    AUTH[Auth TUI]
    REPO[Repo Downloader]
  end

  subgraph EXT["External"]
    APP["../gofast-app<br/>(template repo)"]
    ADMIN["admin.gofast.live<br/>(auth + download)"]
  end

  CMD --> CFG
  CMD --> INT
  CMD --> SVL
  CMD --> E2E
  CMD --> AUTH
  CMD --> REPO
  REPO --> APP
  REPO --> ADMIN
  AUTH --> ADMIN
```

### 2.2 Code generation flow (`gof model`)

```mermaid
sequenceDiagram
  participant U as User
  participant M as model.go
  participant DB as model_db.go
  participant SVC as model_service.go
  participant TST as model_test_gen.go
  participant E as e2e.go
  participant S as client generators

  U->>M: gof model note title:string
  M->>DB: generateProto() + generateSchema() + generateQueries()
  M->>M: generateAuthAccessFlags() + wireCoreMain()
  M->>SVC: generateServiceLayer() + generateTransportLayer()
  M->>TST: generateServiceTestContent() + generateValidationTestContent()
  M->>TST: generateTransportTestContent()
  M->>E: GenerateClientE2ETest() + UpdateSeedDevUser()
  M->>S: Generate client scaffolding for each enabled frontend
```

### 2.3 Integration addition flow (`gof add`)

```mermaid
flowchart TD
  A[gof add stripe] --> B[Authenticate]
  B --> C[Download template to tmpDir]
  C --> D[Copy integration files<br/>domain, transport, migrations]
  D --> E[Merge markers into main.go, config.go]
  E --> F[Strip OTHER integrations' markers]
  F --> G[Add client pages for each enabled frontend]
  G --> H[Update gofast.json]
```

---

## 3. File Tree (Curated) [STABLE]

```text
cmd/gof/
├── main.go                    # Entry point -> cmd.Execute()
├── build.sh                   # Cross-platform build (linux, darwin, windows)
├── cmd/
│   ├── root.go                # Root Cobra command
│   ├── init.go                # gof init - project scaffolding
│   ├── model.go               # gof model - CRUD generation orchestrator (560 lines)
│   ├── model_db.go            # Proto, SQL migration, SQLC query generation
│   ├── model_service.go       # Service + transport + validation generation
│   ├── model_test_gen.go      # Test generation for service/transport/validation (542 lines)
│   ├── add.go                 # gof add - integration dispatcher
│   ├── client.go              # gof client - frontend scaffolding
│   ├── infra.go               # gof infra - Terraform/deployment files
│   ├── mon.go                 # gof mon - monitoring stack
│   ├── auth.go                # gof auth - authentication
│   └── version.go             # gof version
├── config/
│   └── config.go              # gofast.json management (v2.17.0)
├── repo/
│   └── repo.go                # Template repo download (admin.gofast.live)
├── integrations/
│   ├── integrations.go        # Core helpers: strip, copy, merge markers (548 lines)
│   ├── stripe.go              # Stripe: strip, add, client
│   ├── s3.go                  # S3: strip, add, client
│   └── postmark.go            # Postmark: strip, add, client
├── svelte/
│   └── svelte.go              # Svelte page generation per model
├── tanstack/
│   └── tanstack.go            # TanStack page generation per model
├── e2e/
│   └── e2e.go                 # Playwright e2e test generation (294 lines)
└── auth/
    ├── auth.go                # Auth flow runner
    ├── bubble.go              # Bubble Tea TUI components
    └── config.go              # Auth config (token storage)
```

Related files outside `cmd/gof/`:
- `go.mod` - Module: `github.com/gofast-live/gofast-cli/v2`, Go 1.25
- `.github/workflows/lint.yml` - golangci-lint CI
- `web/` - Marketing website (SvelteKit on Cloudflare Workers) - not part of CLI
- `cmd/gofast/` - Legacy v1 CLI - ignore

---

## 4. Core Contracts [STABLE]

### 4.1 CLI commands

| Command | Purpose |
|---------|---------|
| `gof init <name>` | Scaffold new project |
| `gof model <name> <col:type...>` | Generate CRUD model with all layers |
| `gof client svelte` | Add Svelte frontend |
| `gof client tanstack` | Add TanStack frontend |
| `gof add stripe` | Add Stripe payments |
| `gof add s3` | Add S3 file storage |
| `gof add postmark` | Add Postmark email |
| `gof infra` | Add Terraform/deployment files |
| `gof mon` | Add monitoring stack (Grafana, Loki, Tempo, Prometheus) |
| `gof auth` | Authenticate with GoFast |
| `gof version` | Print version (v2.17.0) |

**Prerequisites for `gof init`:** buf, sqlc, goose, docker, docker-compose

### 4.2 Model generation contract

**Syntax:** `gof model <name> <col1:type> <col2:type> ...`

**Model name rules:**
- Lowercase letters and underscores only (e.g., `user_profile`, `event_log`)
- Must be singular - plural names rejected with suggestion (e.g., `trucks` -> use `truck`)
- Minimum 2 columns required

**Column name rules:**
- Lowercase letters, numbers, underscores (e.g., `view_count`, `is_active2`)
- Reserved names rejected: `id`, `user_id`, `created`, `updated` (auto-generated)
- Go keywords rejected: `type`, `func`, `var`, `package`, `map`, `chan`, etc.

**Column types:**

| Type | SQL | Proto | Go | Validation |
|------|-----|-------|-----|------------|
| string | text | string | string | required, minlength(3) |
| number | numeric | string | string | must parse float, >= 1 |
| date | timestamptz | string | time.Time | valid date format (RFC3339 or YYYY-MM-DD) |
| bool | boolean | bool | bool | none |

**Generated artifacts per model:**
1. `proto/v1/{name}.proto` + appends to `main.proto`
2. `app/service-core/storage/migrations/{num}_create_{plural}.sql`
3. Appends queries to `app/service-core/storage/query.sql`
4. `app/service-core/domain/{pkgname}/` - service.go, validation.go, service_test.go, validation_test.go
5. `app/service-core/transport/{pkgname}/` - route.go, route_test.go
6. Wires into `app/service-core/main.go` (imports, deps, routes)
7. Updates `app/pkg/auth/auth.go` (permission flags)
8. Updates `scripts/seed_dev_user.sh` (permission bitmask)
9. `e2e/{plural}.test.ts` (if at least one client exists, since `gof client` owns the `e2e/` folder)
10. Client pages for each configured frontend (Svelte and/or TanStack)

### 4.3 Naming conversions

| Input (snake_case) | Output | Used for |
|---------------------|--------|----------|
| `user_profile` | `UserProfile` (PascalCase) | Go types, proto messages |
| `user_profile` | `userProfile` (camelCase) | Go variables, TS proto field access |
| `user_profile` | `userprofile` | Go package/directory names |
| `user_profile` | `user_profiles` (pluralized) | Table names, route paths, proto service |

### 4.4 Configuration (`gofast.json`)

```json
{
  "project_name": "myapp",
  "services": [
    {"name": "core", "port": "4000"},
    {"name": "svelte", "port": "3000"},
    {"name": "tanstack", "port": "3000"}
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
  "integrations": ["stripe", "s3"],
  "infra_populated": true,
  "monitoring_populated": false
}
```

**Always use config checks, not file existence:**
- `config.HasService("svelte")` / `config.HasService("tanstack")`, not `os.Stat(...)`
- `config.HasIntegration("stripe")` to check integrations

---

## 5. Marker System [STABLE]

### Integration markers

Code in `../gofast-app` is wrapped with markers for optional features. On `gof init`, all integration markers are stripped. On `gof add`, the target integration's markers are kept and others stripped.

**Go files:** `// GF_<INTEGRATION>_START` / `// GF_<INTEGRATION>_END`
**SQL files:** `-- GF_<INTEGRATION>_START` / `-- GF_<INTEGRATION>_END`

| Integration | Marker prefix |
|-------------|---------------|
| Stripe | `GF_STRIPE_` |
| S3/Files | `GF_FILE_` |
| Postmark/Email | `GF_EMAIL_` |

**Files with integration markers:**
- `app/service-core/main.go` - imports, deps, route mounting
- `app/service-core/config/config.go` - integration-specific config fields
- `app/service-core/storage/query.sql` - integration queries
- `app/service-core/storage/migrations/` - integration tables
- `proto/v1/main.proto` - service definitions
- `app/service-core/domain/login/service.go` - `CheckUserAccess()` (Stripe-specific)

### Model wiring markers (in generated project)

**In `app/service-core/main.go`:**
- `GF_MAIN_IMPORT_SERVICES_START/END` - Service imports
- `GF_MAIN_IMPORT_ROUTES_START/END` - Route imports
- `GF_MAIN_INIT_SERVICES_START/END` - Deps initialization
- `GF_MAIN_MOUNT_ROUTES_START/END` - Route mounting

**In `app/service-core/config/config.go`:**
- `GF_CONFIG_STRUCT_INSERT` - Struct fields
- `GF_CONFIG_INIT_INSERT` - Initialization code

**In `app/pkg/auth/auth.go`:**
- `GF_ACCESS_FLAGS_END` - Permission flag constants (insert before)
- `GF_USER_ACCESS_END` - UserAccess bitmask entries (insert before)

### Test generation markers (in skeleton templates)

| Marker | Location | Purpose |
|--------|----------|---------|
| `GF_TP_TEST_ENTITY_FIELDS_START/END` | service_test.go, route_test.go | Database entity field assignments |
| `GF_TP_TEST_CREATE_FIELDS_START/END` | service_test.go, route_test.go | Proto creation request fields |
| `GF_TP_TEST_EDIT_FIELDS_START/END` | service_test.go, route_test.go | Proto edit request fields |
| `GF_TP_TEST_INVALID_FIELDS_START/END` | service_test.go | Invalid proto fields for validation tests |
| `GF_TP_TEST_EDIT_ASSERT_START/END` | route_test.go | Assertions after edit operations |
| `GF_FIXTURES_START/END` | validation_test.go | Proto builder helper functions |
| `GF_MODEL_CONFIG_START/END` | skeletons.test.ts | E2E test model config object |

### How marker replacement works

1. Read skeleton template file
2. Check conditions (has date columns? has non-bool columns?)
3. Build field-specific content per column type
4. Replace each marker region using `replaceMarkerRegion()` - preserves indentation
5. Strip marker lines themselves (clean output)
6. Apply token replacement (`skeleton` -> model name, etc.)
7. Write generated file

---

## 6. Test Generation Details [STABLE]

### Permission bitmask system

Uses bit-shifting (iota pattern) for per-model CRUD permissions:
- Bits 0-1: Plan flags (Free, Pro)
- Bits 2+: Model flags (4 per model: Get, Create, Edit, Remove)
- Final 8 bits: Integration flags (Stripe, S3/Files, Postmark/Email)

`e2e/e2e.go:ComputeUserAccess()` recalculates the dev user permission value dynamically.

Client-side permission marker updates were removed. Svelte and TanStack no longer have frontend permission-marker injection; user access is edited through a simple input in the generated app template.

### Generated Go tests

**Service tests** (`domain/{model}/service_test.go`):
- Test environment with real PostgreSQL (testutil)
- Tests per function: Unauthorized, Forbidden, Validation Error, Success
- Factory helpers: `createTestSkeleton`, `contextWithUser`

**Validation tests** (`domain/{model}/validation_test.go`):
- Table-driven tests for both Create and Edit validation
- Per-column: string (required, minlength), number (parse, gte), date (format)
- Edit adds UUID validation cases
- If all columns are bool, validation error test is removed

**Transport tests** (`transport/{model}/route_test.go`):
- HTTP server with mock auth interceptor
- Tests: Create, GetByID, Edit, Remove, GetAll (streaming)
- Uses ConnectRPC client for type-safe requests

### Generated E2E tests (Playwright)

`e2e/{plural}.test.ts`:
- Create with field validation
- List and verify created item
- Edit and assert updated values
- Delete and verify removal
- Stream get-all endpoint

### Smart conditional behavior

- Removes `"time"` import if no date columns
- Removes Create validation error test if only bool columns
- Generates appropriate test values per column type
- Handles `formatDate` inclusion only when date columns present (Svelte)
- `gof client` copies the full `e2e/` folder as-is, including integration tests
- `gof add stripe|s3|postmark` does not modify `e2e/`
- TanStack formatting runs `npm ci`, regenerates `src/routeTree.gen.ts` with `@tanstack/router-generator`, then formats
- Svelte and TanStack are intentionally aligned at CLI time: no frontend build/typecheck is run by the CLI

---

## 7. Telemetry and Observability [STABLE]

- No telemetry in the CLI itself
- Generated projects include OpenTelemetry tracing in service layer (via `ot.StartSpan`)
- `gof mon` adds Grafana/Loki/Tempo/Prometheus monitoring stack
- CI: golangci-lint on push/PR to main

---

## 8. Current Work [VOLATILE]

- Active initiatives: None documented
- Integration naming: `r2` was renamed to `s3` (more generic, works with any S3-compatible storage)

---

## 9. Future Work [VOLATILE]

1. Additional client frameworks (Next.js, Vue - stubbed in client.go but not implemented)
2. Additional integrations (new marker prefixes)
3. Unit tests for the CLI itself (currently tested only via generated project verification)

Known gaps:
- No automated CI pipeline that generates a demo project and runs its tests
- `scripts/run_tests.sh` referenced in CONTEXT but doesn't exist in CLI repo (lives in generated project)

---

## 10. Critical Invariants and Tricky Flows [STABLE]

### 10.1 Security/scoping invariants
- Authentication required for all commands that download templates (init, add, client, infra, mon)
- `TEST=true` bypasses auth and uses local `../gofast-app` (development only)
- Generated projects scope all queries by `user_id` - never expose other users' data

### 10.2 Data integrity invariants
- Migration numbering must be calculated dynamically from existing files - never hardcoded
- `gofast.json` is the source of truth for enabled features, not file existence
- Permission bitmask must be recalculated whenever models are added

### 10.3 High-risk flows

**Integration addition order:** Adding integrations before/after client affects different code paths. The `gof add` command checks configured frontend services to decide whether to also add client-side pages. If client is added after integration, `gof client` must handle already-enabled integrations.

**E2E ownership:** `gof client` is the single owner of the `e2e/` folder. `gof add` no longer adds/removes e2e files.

**Marker merging:** When adding an integration, the CLI copies files with markers from the template and merges marker blocks into existing project files (main.go, config.go). It must strip markers for OTHER integrations that aren't enabled, while keeping the target integration's markers.

### 10.4 Easy-to-break gotchas
- Proto uses snake_case (`published_at`), TypeScript uses camelCase (`publishedAt`) - Svelte generation must use `toCamelCase()`
- Go package names strip underscores: `user_profile` -> package `userprofile`
- Plural detection uses `go-pluralize` - some edge cases may not pluralize correctly
- Adding client generates pages for ALL existing models in config, not just new ones
- Adding TanStack client to a project with existing models requires route-tree regeneration after route scaffolding; the CLI now does this directly via TanStack's router generator instead of `vite build`
- `model_test_gen.go` has multiple instances of the same marker (e.g., `GF_TP_TEST_CREATE_FIELDS` appears 3+ times in service_test.go) - `replaceMarkerRegion` must handle all occurrences

---

## 11. Quick Reference APIs [STABLE]

```go
// Config
config.ParseConfig() (*Config, error)
config.Initialize(projectName string) error
config.AddModel(name string, columns []Column) error
config.AddIntegration(name string) error
config.HasService(name string) bool
config.AddService(name, port string) error
config.HasIntegration(name string) bool
config.MarkInfraPopulated() error
config.MarkMonitoringPopulated() error

// Integrations
integrations.StripIntegration(projectPath, integration string) error
integrations.RemoveMarkerBlocks(content, startMarker, endMarker string) string
integrations.CopyDir(src, dst string) error
integrations.CopyFile(src, dst string) error
integrations.GetNextMigrationNumber(migrationsDir string) (int, error)
integrations.MergeMainGoMarkers(srcMain, dstMain, integration string) error
integrations.MergeConfigMarkers(srcConfig, dstConfig, integration string) error
integrations.StripOtherIntegrations(projectPath string, keep string) error

// Repo
repo.DownloadRepo(email, apiKey, projectName string) error

// E2E
e2e.GenerateClientE2ETest(modelName string, columns []config.Column) error
e2e.ComputeUserAccess(numModels int) int
e2e.UpdateSeedDevUser() error

// Svelte
svelte.GenerateSvelteScaffolding(modelName string, columns []config.Column) error

// TanStack
tanstack.GenerateTanstackScaffolding(modelName string, columns []config.Column) error

// Naming helpers (in model.go)
toCamelCase(s string) string      // snake_case -> PascalCase
toGoPackageName(s string) string  // snake_case -> lowercase (no underscores)
toGoVarName(s string) string      // snake_case -> camelCase
```

---

## 12. Runbook [VOLATILE]

### 12.1 Local development

```bash
# Run any gof command locally
TEST=true go run ./cmd/gof/... <command> <args>

# Full regeneration test
rm -rf demo
TEST=true go run ./cmd/gof/... init demo
cd demo

# Before `make gen`, edit buf.gen.yaml to use local plugins:
# version: v2
# plugins:
#   - local: protoc-gen-go
#     out: app/gen
#     opt: paths=source_relative
#   - local: protoc-gen-connect-go
#     out: app/gen
#     opt: paths=source_relative
#   - local: protoc-gen-es
#     out: app/service-svelte/src/lib/gen
#     opt: target=ts
#   - local: protoc-gen-es
#     out: app/service-tanstack/src/lib/gen
#     opt: target=ts

TEST=true go run ../cmd/gof/... client svelte
# or: TEST=true go run ../cmd/gof/... client tanstack
TEST=true go run ../cmd/gof/... add stripe
TEST=true go run ../cmd/gof/... add s3
TEST=true go run ../cmd/gof/... add postmark
TEST=true go run ../cmd/gof/... model note title:string content:string views:number published:date active:bool

# Verify generated code compiles and tests pass
docker compose up postgres -d
goose -dir app/service-core/storage/migrations postgres \
  "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" up
cd app/service-core && go test -race ./...
```

### 12.2 Building release binaries

```bash
cd cmd/gof
bash build.sh
# Produces: gof-linux-amd64, gof-darwin-amd64, gof-windows-amd64.exe
```

### 12.3 Debugging checklist
1. Check `gofast.json` - is the config state correct?
2. Check `../gofast-app` - does the template have the expected markers/content?
3. Check generated files in `demo/` - did token replacement work correctly?
4. For test failures: check migration numbering, permission bitmask, import statements
5. For integration issues: verify marker stripping left the right code in place

### 12.4 Generated project Makefile commands

| Command | Description |
|---------|-------------|
| `make start` | Start services with Docker Compose |
| `make starts` | Start with the Svelte client |
| `make startt` | Start with the TanStack client |
| `make startm` | Start with monitoring stack |
| `make keys` | Generate public/private keys |
| `make sql` | Regenerate SQLC queries |
| `make gen` | Regenerate proto/buf code |
| `make migrate` | Apply database migrations |

---

## 13. Infrastructure [STABLE]

**Approach: Static configuration with empty defaults.**

Infrastructure files in `gofast-app` come with all integrations pre-configured. Integration-specific variables default to empty strings, so unused integrations don't break anything.

**Key infra files:**
- `infra/integrations.tf` - All integration variables (default = "")
- `infra/variables.tf` - Base infrastructure variables
- `infra/secrets.tf` - Base secrets (regcred, cron, e2e, google-oauth)
- `infra/service-core.tf` - Deployment with all integration env vars pre-wired

No dynamic injection for infra files - they're copied as-is from the template.

---

## Dependencies [STABLE]

| Package | Purpose |
|---------|---------|
| `github.com/spf13/cobra` v1.9.1 | CLI framework |
| `github.com/charmbracelet/bubbletea` v0.26.6 | Terminal UI for auth |
| `github.com/charmbracelet/bubbles` v0.18.0 | TUI components |
| `github.com/charmbracelet/lipgloss` v0.12.1 | Terminal styling |
| `github.com/gertd/go-pluralize` v0.2.1 | Model name pluralization |
