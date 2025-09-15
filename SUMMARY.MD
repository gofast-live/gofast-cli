# GoFast CLI (`gof`) Summary for LLMs

This document provides a high-level overview of the GoFast CLI (`gof`), a tool designed for scaffolding full-stack Go applications. The goal is to enable rapid development by automating the creation of boilerplate code across the entire stack, from the database to the frontend.

## Core Philosophy

The CLI follows a "Lego brick" approach. A developer starts with a foundational service and then adds features (data models, frontend clients) incrementally. The CLI handles the "plumbing" for each new piece, ensuring all layers of the application are aware of the new feature and can interact with it.

## Key Commands and Workflow

The typical workflow involves using a sequence of commands to build an application.

### 1. `gof auth`
- **Purpose**: Authenticates the user with the GoFast backend service.
- **Mechanism**: Prompts for an email and API key, validates them against the server, and stores them in a local config file (`~/.config/gofast.json`). This is a prerequisite for commands that need to download templates.

### 2. `gof init [project_name]`
- **Purpose**: Initializes a new GoFast project structure.
- **Actions**:
    - Downloads a template repository.
    - Sets up a Go-based backend service (`service-core`).
    - Configures Docker (`docker-compose.yml`) for a PostgreSQL database.
    - Creates a `gofast.json` file to track project configuration (services, models).
    - Runs initial setup scripts for generating keys, database schemas, and gRPC/Protobuf code.
    - The initial template includes a `skeleton` model as a reference.

### 3. `gof client [client_type]`
- **Purpose**: Adds a frontend application to the project.
- **Supported Types**: `svelte` (fully implemented), `next`, `vue` (planned).
- **Actions**:
    - Downloads the specified client template into the `app/service-client` directory.
    - Updates `docker-compose.yml` and `buf.gen.yaml` to integrate the client into the project's build and runtime environment.
    - If data models already exist in `gofast.json`, it automatically generates the necessary UI pages (list, add, edit) for each model.

### 4. `gof model [model_name] [columns...]`
- **Purpose**: The core scaffolding engine of the CLI. It creates all the necessary code for a new data model across every layer of the application.
- **Arguments**:
    - `model_name`: The singular name of the model (e.g., `post`).
    - `columns`: A list of columns defined as `name:type`. Valid types are `string`, `number`, `time`, and `bool`.
- **End-to-End Code Generation**: The generation process is template-based. For each layer (domain, transport, UI), there is a `skeleton` directory containing template files. The CLI copies the relevant skeleton and performs find-and-replace operations for the model name and its columns. To facilitate dynamic content injection, the templates use special `GF_` prefixed comment markers (e.g., `// GF_FIXTURES_START`, `<!-- GF_DETAIL_FIELDS_END -->`) to delineate regions that should be programmatically updated. This process includes:
    1.  **Protobuf (`/proto/v1`)**: Generates a new `.proto` file for the model and adds the corresponding service definitions to `main.proto`.
    2.  **Database (`/app/service-core/storage`)**:
        - Appends a `CREATE TABLE` statement to `schema.sql`.
        - Appends type-safe SQLC queries (CRUD operations) to `query.sql`.
    3.  **Permissions (`/app/pkg/auth`)**: Adds access control flags for the new model (e.g., `CreatePost`, `EditPost`).
    4.  **Domain Layer (`/app/service-core/domain/[model_name]`)**: Creates the business logic, including:
        - A `service.go` file for core logic.
        - A `validation.go` file with request validation rules.
        - Corresponding `_test.go` files for the service and validation logic.
    5.  **Transport Layer (`/app/service-core/transport/[model_name]`)**: Creates the gRPC/HTTP handlers (`route.go`) that expose the domain logic as API endpoints, along with tests.
    6.  **Wiring**:
        - Injects the new model's service and routes into the main server (`server.go`).
        - Initializes the service in `main.go`.
    7.  **Frontend (`/app/service-client`)**: If a client is installed, it generates:
        - A list page to display all items of the model.
        - A detail page for creating and editing items.
        - Unit tests (`.spec.ts`) for the new pages.
    8.  **Build Scripts**: Runs `sqlc` and `buf` to generate Go code from the new SQL queries and Protobuf definitions.

## Summary

The `gof` CLI is a powerful code generation tool that automates the creation of a full-stack, production-ready Go application. Its primary strength lies in the `gof model` command, which builds out an entire vertical slice of functionality—from database to UI—with a single command, significantly accelerating development time.
