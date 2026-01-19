# GoFast CLI

Building blocks for Go.
Generate production-ready Go apps with ConnectRPC, SvelteKit, and PostgreSQL.

This repository contains two CLI tools:
- **`gofast`** - v1 CLI (legacy)
- **`gof`** - v2 CLI (current)

---

## GoFast CLI v2 (`gof`)

The v2 CLI generates full-stack Go applications with:
- Go backend with ConnectRPC transport
- PostgreSQL database with SQLC
- OAuth authentication (GitHub, Google, Phone)
- Optional Svelte frontend
- Optional integrations (Stripe, R2, Postmark)

### Installation

#### Using Go (Recommended)

```bash
go install github.com/gofast-live/gofast-cli/v2/cmd/gof@latest
```

Make sure your `PATH` includes the Go bin directory:

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

#### Download Binary

Go to the [Releases](https://github.com/gofast-live/gofast-cli/releases) page and download the `gof` binary for your OS.

### Prerequisites

Get your API key at [admin.gofast.live](https://admin.gofast.live) ($40 one-time purchase).

### Quick Start

```bash
# 1. Authenticate (one-time - requires API key)
gof auth

# 2. Create a new project
gof init myapp
cd myapp

# 3. Add models
gof model note title:string content:string
gof model task title:string done:bool due_date:date

# 4. Run code generation
make sql       # Generate SQLC queries
make gen       # Generate proto code
make migrate   # Apply database migrations

# 5. Start the app
make start
```

### Commands

| Command | Description |
|---------|-------------|
| `gof auth` | Authenticate with GoFast |
| `gof init <name>` | Create new project |
| `gof model <name> [cols...]` | Generate CRUD model |
| `gof client svelte` | Add Svelte frontend |
| `gof add stripe` | Add Stripe payments |
| `gof add r2` | Add Cloudflare R2 storage |
| `gof add postmark` | Add Postmark email |
| `gof infra` | Add Terraform/deployment files |
| `gof version` | Show CLI version |

### Model Column Types

```bash
gof model post title:string views:number published_at:date is_active:bool
```

| Type | SQL | Example |
|------|-----|---------|
| `string` | text | `title:string` |
| `number` | numeric | `views:number` |
| `date` | timestamptz | `published_at:date` |
| `bool` | boolean | `is_active:bool` |

### Example Workflow

```bash
# Create project
gof init blog
cd blog

# Add models
gof model post title:string body:string published:bool
gof model comment content:string

# Add frontend
gof client svelte

# Add payments
gof add stripe

# Add infrastructure/monitoring
gof infra

# Generate code
make sql && make gen && make format && make migrate

# Run with client
make startc

# Run with client + monitoring (Grafana, Alloy, Loki, Tempo, Prometheus)
make startcm
```

### Generated Project Commands

| Command | Description |
|---------|-------------|
| `make start` | Start backend services |
| `make startc` | Start with Svelte client |
| `make startm` | Start with monitoring (Grafana, Alloy, Loki, Tempo, Prometheus) |
| `make startcm` | Start with client + monitoring |
| `make sql` | Regenerate SQLC queries |
| `make gen` | Regenerate proto code |
| `make migrate` | Apply database migrations |
| `make format` | Format all code |

---

## GoFast CLI v1 (`gofast`)

Legacy CLI. See [docs.gofast.live](https://docs.gofast.live) for v1 documentation.

### Installation

```bash
go install github.com/gofast-live/gofast-cli/v2/cmd/gofast@latest
```

---

## Building from Source

```bash
git clone https://github.com/gofast-live/gofast-cli.git
cd gofast-cli

# Build v2 (gof)
go build -o gof ./cmd/gof/...

# Build v1 (gofast)
go build -o gofast ./cmd/gofast/...

# Cross-compile v2
GOOS=linux GOARCH=amd64 go build -o gof-linux-amd64 ./cmd/gof/...
GOOS=darwin GOARCH=amd64 go build -o gof-darwin-amd64 ./cmd/gof/...
GOOS=windows GOARCH=amd64 go build -o gof-windows-amd64.exe ./cmd/gof/...
```

## License

MIT License - see [LICENSE](LICENSE) for details.
