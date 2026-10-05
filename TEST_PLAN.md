# Test plan: RustFS local S3, Delve debugging, API key auth

Temporary handoff file. Delete it once testing is done, don't commit it.

Branch `feat/minio-delve-apikey` exists in both `gofast-app` (template) and `gofast-cli`.
Nothing is committed. Read `CONTEXT.md` in both repos first. This plan only lists what is new.

## What changed

**Template (`gofast-app`)**
1. **Local S3 (the "MinIO integration" request).** MinIO no longer publishes images: `minio/minio`
   on Docker Hub is gone and quay.io denies access. `docker-compose.yml` now runs
   `rustfs/rustfs:1.0.0`, a MinIO-compatible server, plus a one-shot `rustfs-init` that creates the
   bucket `files` with a SigV4-signed curl. Both are inside a `# GF_FILE_START/END` block at the end of
   the file. The core's `S3_*` defaults point at `http://rustfs:9000`.
   Root bug fix: `domain/file/s3.go` now sets `UsePathStyle = true`. Without it, every self-hosted
   S3 got `NoSuchBucket` (virtual-host addressing).
2. **Delve.** `docker-compose.debug.yml` overlay plus `app/service-core/.air.debug.toml`, which
   builds with `-N -l` and runs `dlv exec --headless :2345 --continue`. `dlv` is installed in the
   Dockerfile `dev` stage. New targets: `make startd`, `startsd`, `starttd`.
3. **API key auth.**
   - `Authorization: Bearer <key>` works on ConnectRPC unary and streaming calls and on the
     `AuthMiddleware` upload routes. The header wins over cookies, and nothing is refreshed or set.
   - Keys are stored as `api_key_hash` (sha256 hex) plus `api_key_hint` (last 4 characters).
     Migration `00006_hash_user_api_keys.sql` converts existing plaintext keys, so they keep working.
   - `GenerateAPIKey` only works for the caller's own id. New users start with no key.
   - `RefreshResponse` gained `id`. Both frontends store it and only show "Generate New API Key" on
     your own user page.
   - New metric `gofast_auth_api_key_total{result,reason}`, with a dashboard panel.

**CLI (`gofast-cli`)**
- `StripIntegration` also strips `# GF_<X>_START/END` blocks in `docker-compose.yml`, so init removes RustFS.
- New `AppendComposeBlock`: `gof add s3` appends the template's `GF_FILE` compose block and renames
  `gofast` to the project name inside that block only.
- Updated the `gof add s3` help text and next-steps output.
- Debug files, migration 00006 and the API key code need no CLI logic. They ship through the
  template copy.
- **Strict lint refactor (451 to 0 issues).** `.golangci.yml` now matches the template's strict config.
  - Cobra commands are now `newXCmd()` constructors, wired in `Execute()`.
  - `init`, `client`, `model` and `mon` were split into step functions.
  - Errors are wrapped, `ctx` is threaded into exec and HTTP calls, and styles are functions.
  - The zip extract now rejects `../` entries and caps each file at 512 MB.
  - A golden diff (old vs new code on the same template copy) gave byte-identical output for
    `model` generation (schema, queries, access flags, service, transport, `main.go`, Svelte and
    TanStack pages, e2e) and for strip plus add of every integration. The only diff was the
    intended RustFS compose rename.
  - Not covered by the golden diff:
    - command output text. A few messages changed, e.g. model column errors now read
      "Error: invalid ..." and client folder errors gained an "Error " prefix.
    - `gof auth` (bubbletea receivers changed, 30s request timeout added).
    - `gof infra`.
    - the real zip download.

4. **User management permissions fix (pre-existing privilege escalation).** The default
   `UserAccess` used to include `GetUsers`, `UpdateUserAccess` and `DeleteUser`, so any user could
   make themselves admin (-1) or delete anyone. Now:
   - Those flags are admin-only.
   - `GetUserByID` and `RemoveUser` allow your own id without them (`authorizeSelfOr`).
   - Frontends: `canManageUsers(access)`. Non-admins on `/users` are redirected to `/users/<own id>`,
     and the Access editor and "Back to Users" are hidden. Deleting your own account sends you to `/login`.
   - `POST /auth/test-login` now requires `admin` (bool). e2e `loginAs(context, email, admin)`:
     `users.test.ts` logs in as admin, every other test as a regular user.
   - New e2e test: "Users Management - regular user › should only manage their own account".
   - The CLI needs no change: `gof model` appends model flags to the last line of `UserAccess`,
     which still works.

5. **Dependency upgrade.**
   - Go 1.26.3 to 1.27.1 (mise, every `go.mod`, `go.work`, Dockerfiles, CI). golangci-lint 2.11.4 to 2.14.0.
     All Go modules got `go get -u` (minor/patch only). otel/log v0.22 moved `Record` to
     `attribute.Value`/`KeyValue`, so `pkg/logger` was adapted.
   - Svelte: Vite 8, vite-plugin-svelte 7, ESLint 10, TypeScript 6.0 (not 7: kit, svelte-check and
     typescript-eslint cap at 6), prettier-plugin-svelte 4, `@lucide/svelte` 1.x. SvelteKit stays on
     2.70 (3.0 is still RC). The `svelte/no-navigation-without-resolve` override is gone, the rule passes.
   - TanStack: Vite 8, plugin-react 6, React 19.3, Start 1.168, `nitro` 3.0.260903-beta (was
     `nitro-nightly`), ESLint 10, TypeScript 6.0. `vite-tsconfig-paths` removed (the `@` alias already
     lives in `resolve.alias`), and `baseUrl` dropped from tsconfig.
   - Lucide 1.x removed brand icons, so the GitHub login icon is now inline SVG in both frontends.
   - Node 24.15.0 to 24.21.0 (mise, Dockerfiles, CI, `engines`; npm floor 11.19.0). Playwright 1.55 to 1.63 (e2e needs `npx playwright install chromium` once locally).

## Already verified (template only, this session)
- Targeted Go tests pass: `pkg/auth`, `pkg/otel`, `domain/login`, `domain/user`, `domain/file`,
  `transport`, `transport/login`, `transport/user`.
- Svelte `npm run check` and TanStack `npm run typecheck` both pass.
- Migration 00006 checked on real Postgres: up to 5, insert key `abc`, then up gives
  `api_key_hash = sha256("abc")` (matches `auth.HashAPIKey`) and hint `abc`. Down works too.
- `UsePathStyle` checked against RustFS: before the fix `NoSuchBucket`, after the fix upload,
  download and delete all succeed.
- `make startd` equivalent: `/health` returns 200, a breakpoint set via Delve RPC was hit, the
  request finished after resume, and an air rebuild restarted dlv.
- Live API key smoke test (curl against the running stack):
  - Key from a cookie session works on `LoginService/Refresh` over Bearer, with no Set-Cookie.
  - A bad key gets 401.
  - Upload, list, download and delete of a file through RustFS all work with the key.
  - `GenerateAPIKey` for another id gets 403.
- Permissions fix: new domain tests (default user can't list users or grant themselves admin; self
  read and self delete allowed) fail with the old `UserAccess` and pass with the new one. Svelte check
  and lint, TanStack typecheck and lint, and `playwright test --list` all compile. Nothing in the UI
  was run in a browser.
- CLI: `StripIntegration` removes the compose block. `AppendComposeBlock` restores it identically
  (with the renamed project), is idempotent, and `docker compose config -q` passes.

- Upgrade: `go build` and `go vet` on all modules, tests in `pkg/...`, `transport/...`,
  `domain/file`, `domain/login`. Svelte `check`, `lint`, `build`. TanStack `check` (lint + typecheck), `build`.
  Not run: Docker image builds on `golang:1.27.1-bookworm`, Delve under Go 1.27, golangci-lint 2.14,
  dev servers in the browser.

NOT verified yet: anything through a real `gof` run, R2, the full Go suite, e2e, lint, monitoring.

## Environment gotchas
- **`goose` on PATH is Block's goose AI agent (mise shim), not pressly/goose.** `make migrate` and
  `gof init` (which runs `make migrate`) break on this machine. Ask the user to fix mise. Until
  then, use `go run github.com/pressly/goose/v3/cmd/goose@v3.28.0 -dir app/service-core/storage/migrations postgres "host=localhost user=postgres password=postgres dbname=postgres sslmode=disable" up`.
  Note that `gof init` will likely fail at "Applying database migrations".
- `buf.gen.yaml` uses remote plugins. Remote worked this session. If not, switch to local plugins
  (see CLI CONTEXT.md step 1.5).
- `app/service-core/tmp/` in the template is root-owned (left by docker air builds). Don't create files there.
- Docker container names are fixed (`<project>-core`, `<project>-rustfs`, ...). Run `make stop` or
  `docker compose down` between scenarios.

## Scenarios (from gofast-cli root, `rm -rf demo` before each)

### 1. Init only, no integrations
```bash
TEST=true go run ./cmd/gof/... init demo && cd demo
```
- [ ] `docker-compose.yml` has no `GF_FILE`, `rustfs:` service or `rustfs-init`. `docker compose config -q` passes.
- [ ] `docker-compose.debug.yml` and `app/service-core/.air.debug.toml` exist. `migrations/` contains `00006_hash_user_api_keys.sql`.
- [ ] `cd app/service-core && go build ./... && go test -race ./...` passes. This covers the API key
      tests with the non-Stripe `CheckUserAccess` fallback, plus `transport/auth_interceptor_test.go`.
- [ ] `make startd`: health 200, `nc localhost 2345` accepts, attaching from VS Code with the README
      `launch.json` hits a breakpoint, and editing a `.go` file triggers a rebuild while dlv comes back.

### 2. Init + `gof add s3`
```bash
TEST=true go run ../cmd/gof/... add s3 && make gen && make sql && make migrate
```
- [ ] The compose file ends with the `GF_FILE` block, and container names are `demo-rustfs` / `demo-rustfs-init`.
- [ ] The files migration is numbered after 00006 (expect `00007_create_files.sql`).
- [ ] `make start`: `rustfs-init` exits 0 (`docker inspect -f '{{.State.ExitCode}}' demo-rustfs-init`)
      and the console at http://localhost:9001/rustfs/console/ loads.
- [ ] `go test -race ./...` passes.
- [ ] Curl smoke (needs `E2E_TEST_SECRET`, default `super-secret`):
      1. Get cookies with `POST /auth/test-login` (header `X-E2e-Test-Secret`), then
         `LoginService/Refresh` returns `id`.
      2. `UserService/GenerateAPIKey {id}` returns a 64-hex key.
      3. With `Authorization: Bearer <key>`:
         - `POST /files/upload -F files=@x.txt` returns 200.
         - `FileService/GetFiles` lists the file.
         - `DownloadFile` returns the same bytes.
         - `RemoveFile` returns 200.
      4. A bad key returns 401 and the response has no `Set-Cookie`.
- [ ] Running `gof add s3` a second time doesn't duplicate the compose block (or is refused by the config check).

### 3. Full stack, Svelte: client + all integrations + a model, then e2e
`gof client svelte`, `gof add stripe`, `gof add s3`, `gof add postmark`, then `gof model note title:string count:number`.
- [ ] Run e2e as in CLI CONTEXT.md "Test Strategy" step 3. Don't export `S3_*` before compose up.
      Pass the RustFS values only to Playwright so `files.test.ts` stops being skipped.
- [ ] `users.test.ts` passes, including "should generate new API key" (own user page) and the new
      "regular user" test.
- [ ] All other e2e files pass with `loginAs(..., false)`. Payments plan bits must be unaffected:
      only `users.test.ts` creates admins.
- [ ] Manual UI as a regular user:
      - The "Users" nav leads to your own page.
      - No Access editor.
      - Deleting your own account lands on `/login`, and logging in again creates a fresh user.
- [ ] Manual UI as an admin: the list works, editing access on another user works, and deleting
      another user returns to `/users`.
- [ ] curl: a default user calling `UserService/UpdateUserAccess {id: self, access: -1}` gets 403,
      and so does `GetAllUsers`.
- [ ] Manual UI: log in as a second user (test-login with another email) and open the first user's
      `/users/<id>`. Expect "Only the owner of this account can generate its API key." and no
      Generate button. On your own page the Generate button shows, plus the
      `Authorization: Bearer` hint.
- [ ] With Stripe: an API key request for a user with an active subscription returns plan bits
      in `Refresh.access` (same value as the cookie session).

### 4. Same as 3 with TanStack, varying the order
- [ ] Client at start, client in the middle, client at end (see the CLI CONTEXT order matrix).
      Watch the `id` in `store.ts` / `_layout.tsx` and the user page.

### 5. Project name edge cases
- [ ] `init gofast-demo` then `add s3`: the compose file must not end up with `gofast-demo-demo-*` names.
- [ ] `init MyDemo` then `add s3`: the bucket is still `files` (fixed on purpose, since S3 bucket names
      must be lowercase) and `rustfs-init` exits 0.

### 6. R2 with path-style (ask user for R2 creds)
- [ ] Export `S3_ACCESS_KEY_ID`, `S3_SECRET_ACCESS_KEY`, `S3_ENDPOINT=https://<account>.r2.cloudflarestorage.com`
      and `BUCKET_NAME` before `make start`, then repeat the curl upload, download and delete.
      `UsePathStyle = true` is now forced for every provider, and R2 is the production target.

### 7. Monitoring
- [ ] `gof mon`, `make startm`, send a few good and bad Bearer requests. Then check that
      `gofast_auth_api_key_total` shows up in VictoriaMetrics (:8428) and that the
      "API Key Auth Outcomes" panel renders in Grafana (:3001).

### 8. Upgrade checks
- [ ] `make starts` and `make startt`: both dev servers come up in Docker (`npm ci` with the new lockfiles,
      Vite 8), pages render, HMR works, login page shows the GitHub icon.
- [ ] `make startd` builds the core on `golang:1.27.1-bookworm` and Delve 1.27.2 hits a breakpoint.
- [ ] Prod images: `docker build --target prod` for core, svelte and tanstack (`.output/server/index.mjs`).
- [ ] `gof model` and `gof client tanstack` output still passes `npm run check` (routeTree regenerates).

### 9. CLI refactor smoke
- [ ] `gof auth` with a real key: TUI focus switching (tab), submit, cancel (esc), and a wrong key shows the error.
- [ ] Real (non-TEST) `gof init` once: zip download and extract still work.
- [ ] `gof infra` and `gof mon` on an existing project, run twice: the second run prints "already exists. Skipping copy."
- [ ] `gof model` with bad input (plural name, `select:string`, a duplicate column, one column) prints a clear error.

### 10. Gates
- [ ] `golangci-lint run` in the template `app/service-core` and in the CLI.
- [ ] `npm run lint` in both frontends.

## Known risks to watch
- The core has no `depends_on` on rustfs, so an upload in the first seconds after `up` can race
  bucket creation. Acceptable for dev; note it if it bites.
- `rustfs-init` uses `curl -sf`. RustFS returns 200 when the bucket already exists. A provider
  answering 409 would make init exit 1.
- Projects generated before this change keep `S3_*` defaults set to empty. `gof add s3` adds the
  RustFS services but doesn't rewrite core env defaults.
- The real `gof` (non-TEST) downloads the published template, so users only get this after gofast-app is released.

## Pre-existing issues found (not fixed, report only)
- `docker-compose.test.yml` overrides a `client` service that no longer exists (left over from service-client).
- CLI `e2e.UpdateSeedDevUser` targets `scripts/seed_dev_user.sh`, which the template no longer has, so it silently no-ops.
