# Project Context: sda (Simple Docker Apps)

## Overview

`sda` is a Go-based CLI tool designed to simplify the process of spinning up development servers (databases, caches, etc.) using Docker and Docker Compose. It provides a unified interface for managing both standalone Docker containers and multi-container Compose stacks, allowing users to define services in a configuration file and manage them with simple, consistent commands.

The tool is intended for local development to avoid repetitive `docker run` commands, manual volume/port configuration, and the complexity of managing both Docker and Compose services separately.

## Key Features

* **Unified Interface:** Same commands work for both Docker and Compose services - users don't need to know the difference
* **Service Configuration:** YAML-based config file with predefined services
* **Compose Integration:** First-class support for Docker Compose files via the `docker compose` CLI
* **Bulk Operations:** Operate on multiple services at once (`--all`, `--running`, `--stopped`)
* **Customization:** Override config settings via CLI flags
* **Interactive Prompts:** Confirmation prompts for destructive operations
* **JSON Output:** All commands support `--json` flag for scripting

## Key Technologies

* **Language:** Go (1.25.0)
* **CLI Framework:** [Cobra](https://github.com/spf13/cobra)
* **Configuration:** [Viper](https://github.com/spf13/viper)
* **Container Runtime:** Docker CLI, invoked via `os/exec` (no Docker SDK dependency - sda is a thin wrapper and assumes `docker` is already installed)
* **Compose Runtime:** `docker compose` CLI plugin, invoked the same way
* **Task Runner:** [Task](https://taskfile.dev/)
* **Interactive Prompts:** [Promptkit](https://github.com/erikgeiser/promptkit)

## Architecture

### Entry Point

* `main.go` calls `cmd.Execute()`

### CLI Commands

Defined in `cmd/` directory:

* `root.go` - Root command, config initialization, version info
* `create.go` - Create/start services (acts as `docker create` + `docker compose up`)
* `start.go` - Start services (acts as `docker start` + `docker compose start`)
* `stop.go` - Stop services (acts as `docker stop` + `docker compose stop`)
* `remove.go` - Remove services (acts as `docker rm` + `docker compose down`)
* `list.go` - List services with filtering options
* `logs.go` - View service logs
* `show.go` - Show service details (Docker only)
* `connect.go` - Connect to services (Docker only)

### Configuration Layer

Located in `internal/config/`:

* **Core Structs:** `Config`, `Service`, `Docker`
* **Service Types:** Services can be either Docker containers OR Compose stacks
  * `Service.Compose` field determines type
  * `Service.IsComposeService()` method for detection
* **Config File:** `~/.config/sda/sda.yaml`
* **Default Config:** Embedded in `cmd/defaultConfig.yaml`

### Docker Integration

Located in `internal/docker/`:

* `api.go` - Resolves the `docker` binary on `PATH` (`exec.LookPath`); `New(cfg)` stores the binary path and `*config.Config` on `Api`
* `exec.go` - The only place that shells out: `capture` (silent, returns stdout, error includes stderr), `run` (stdout/stderr attached, for pull/compose progress), `runInteractive` (+ stdin, for `docker exec -it`)
* `create.go` - Builds `docker create` argv from a `config.Service` (`buildCreateArgs`), pulls the image first if missing
* `compose.go` - Wraps the `docker compose` CLI
  * `ComposeUp()` - `docker compose up --detach --remove-orphans [--build] [--force-recreate]`
  * `ComposeStart()` - `docker compose start`
  * `ComposeStop()` - `docker compose stop`
  * `ComposeDown()` - `docker compose down --remove-orphans [--volumes]`
  * `ComposeLogs()` - `docker compose logs [--follow]`
  * Path resolution (relative/absolute, file/directory)
  * Folder name validation (must match service name)
* `operations.go` - List, Start, Stop, Remove, Connect, GetInfo, Logs operations, all shelling out to `docker ps`/`start`/`stop`/`rm`/`volume rm`/`exec`/`logs`
* `network.go` - Network management via `docker network inspect`/`create`
* `transformations.go` - Helper transformations (ports, versions, etc.)
* `shellwords.go` - `splitArgs()`, a small quote-aware tokenizer used to turn a `cliConnectCommand` string into `docker exec` argv
* `types.go` - Type definitions (`ServiceInfo`, `LogsOptions`)

### Utilities

Located in `internal/utils/`:

* `output.go` - Message, Error, and JSON output functions
* `prompts.go` - Confirmation prompts
* `commands.go` - `OpenURL` (opens a browser for web-connect services)
* `customFlags.go` - Custom flag types (Enum)

### Build Tools

Located in `bin/`:

* `gendocs.go` - Standalone documentation generator
  * Uses Cobra's doc generation
  * Supports `-output` flag
  * Generates markdown and man pages
  * NOT included in shipped binary

### Testing

Located in `test/` and `internal/*/`:

* Config tests: 94.4% coverage
* Docker transformations tests: 100% coverage on core functions
* Utils tests: Custom flag validation
* Test helpers for temp config and directory creation

## Build and Run

### Prerequisites

* Go 1.22 or higher
* Docker CLI ≥ 23.0 (running) - sda shells out to `docker`, it does not vendor a Docker SDK
* Docker Compose v2 plugin (`docker compose`), for compose services
* `task` (optional, for convenience)

### Commands (using `task`)

The project uses a `Taskfile.yaml` for common tasks:

* **Build:** `task build` (outputs to `publish/sda.exe` or `publish/sda`)
* **Run:** `task run -- [command]` (runs the built binary)
* **Build and Run:** `task runb` (builds then runs)
* **Test:** `task test` (runs `go test -v ./...` with coverage)
* **Build All Platforms:** `task build-all` (cross-compiles for Windows/Linux/Darwin AMD64/ARM64)
* **Generate Docs:** `task docs` (runs `bin/gendocs.go`)
* **Clean:** `task clean`
* **Version Increment:** `task version:increment [patch|minor|major]` (creates git tag)

### Commands (standard Go)

* **Build:** `go build -o publish/sda main.go`
* **Run:** `go run main.go [command]`
* **Test:** `go test -v ./...`
* **Generate Docs:** `go run bin/gendocs.go [-output docs]`

## Configuration

### Config File Location

`$HOME/.config/sda/sda.yaml`

### Config Structure

```yaml
defaultNetwork: sda-network
defaultPassword: password
prefix: sda-

services:
  # Docker service example
  - name: postgres
    outputName: PostgreSQL
    defaultVersion: latest
    hasPassword: true
    docker:
      imageName: postgres
      portMappings:
        - host: 5432
          container: 5432
      volumes:
        - source: postgres-data
          target: /var/lib/postgresql/data
          isNamed: true
      envVars:
        - POSTGRES_PASSWORD={{password}}
    hasCliConnect: true
    cliConnectCommand: psql -U postgres
    hasWebConnect: false

  # Compose service example
  - name: dagu
    outputName: Dagu
    compose: ./dagu/docker-compose.yml  # Relative to ~/.config/sda/
```

### Service Types

**Docker Service:**

* Has `docker` section with image, ports, volumes, env vars
* Managed as individual containers with `sda-` prefix
* Full customization via CLI flags

**Compose Service:**

* Has `compose` field pointing to docker-compose file
* Managed as compose stack using service name as project name
* Path can be:
  * Relative to `~/.config/sda/` (e.g., `./myapp/docker-compose.yml`)
  * Absolute path (e.g., `/home/user/projects/app/docker-compose.yml`)
  * Directory (searches for `docker-compose.yaml` or `docker-compose.yml`)
  * File path (used directly)
* **Validation:** Folder containing compose file MUST match service name

### Config Initialization

* On first run, if config doesn't exist, creates default config from embedded `cmd/defaultConfig.yaml`
* Users can specify custom config with `--config` flag

## Command Reference

### Unified Commands (Work with Both Service Types)

**create** - Create and start service

```bash
sda create [service] [flags]
  --build           # Build images before starting (compose only)
  --recreate        # Recreate if exists
  --volumes         # Remove volumes when recreating
  --no-start        # Create but don't start
  --port HOST:CONTAINER  # Override port mapping (docker only)
  --volume SRC:TGT  # Override volume mapping (docker only)
  --env KEY=VALUE   # Override environment variable (docker only)
  --network NAME    # Override network (docker only)
  --password PASS   # Override password (docker only)
  --version VER     # Override version (docker only)
```

**start** - Start existing service

```bash
sda start [service|--all|--running|--stopped]
```

**stop** - Stop running service

```bash
sda stop [service|--all|--running|--stopped]
```

**remove** - Remove service

```bash
sda remove [service|--all|--running|--stopped]
  --volumes         # Also remove volumes
  -y, --yes         # Skip confirmation
```

**logs** - View service logs

```bash
sda logs [service]
  -f, --follow      # Follow log output
  --tail N          # Show last N lines
  --timestamps      # Show timestamps
```

**list** - List services

```bash
sda list [flags]
  -a, --available   # List all available services
  -c, --created     # List created services
  -r, --running     # List running services (default)
  -s, --stopped     # List stopped services
  --compose         # List only compose services
  --no-color        # Disable colored output
  -f, --format FORMAT  # Output format (table|json)
```

### Docker-Only Commands

**show** - Show service information

```bash
sda show [service]
```

**connect** - Connect to service

```bash
sda connect [service]
  --password PASS   # Override password
  --web             # Force web connection
```

### Global Flags

* `--config string` - Custom config file path
* `--json` - Output as JSON
* `-y, --yes` - Skip confirmation prompts

## Development Conventions

### Project Structure

```text
sda/
├── cmd/                    # CLI commands
│   ├── root.go            # Root command, config init
│   ├── create.go          # Create command
│   ├── start.go           # Start command
│   ├── stop.go            # Stop command
│   ├── remove.go          # Remove command
│   ├── list.go            # List command
│   ├── logs.go            # Logs command
│   ├── show.go            # Show command
│   └── connect.go         # Connect command
├── internal/
│   ├── config/            # Configuration management
│   │   └── config.go      # Config structs and methods
│   ├── docker/            # Docker/Compose operations (shells out to the `docker` CLI)
│   │   ├── api.go         # Resolves docker on PATH; New(cfg) stores *config.Config
│   │   ├── exec.go        # capture/run/runInteractive - the only os/exec call sites
│   │   ├── create.go      # docker create argv builder + image pull
│   │   ├── compose.go     # docker compose CLI wrapper
│   │   ├── operations.go  # CRUD + list/logs/connect operations
│   │   ├── network.go     # Network management
│   │   ├── transformations.go  # Helper functions
│   │   ├── shellwords.go  # cliConnectCommand -> argv tokenizer
│   │   └── types.go       # Type definitions
│   └── utils/             # Utility functions
│       ├── output.go      # Output formatting
│       ├── prompts.go     # User prompts
│       ├── commands.go    # OpenURL (browser launch)
│       └── customFlags.go # Custom flag types
├── bin/                   # Build-time tools
│   └── gendocs.go         # Documentation generator
├── test/                  # Test files
│   └── helpers.go         # Test helpers
├── _dev/                  # Development planning docs
│   ├── TODO.md            # Remaining tasks
│   ├── PLAN.md            # Execution plan
│   └── NEXT.md            # Next phase details
├── publish/               # Build output (gitignored)
├── docs/                  # Generated docs (gitignored)
├── main.go                # Entry point
├── Taskfile.yaml          # Task runner config
└── go.mod                 # Go module definition
```

### Code Patterns

**Service Type Detection:**

```go
service := config.CONFIG.GetServiceByName(name)
if service != nil && service.IsComposeService() {
    // Handle as compose service
    client.ComposeUp(*service, build, recreate)
} else {
    // Handle as Docker service
    client.Start(name)
}
```

**Error Handling:**

```go
if err != nil {
    utils.Error(fmt.Sprintf("Error message: %v", err))
    utils.ErrorAndExit("")
}
```

**Confirmation Prompts:**

```go
if !yes {
    if !utils.Confirm("Are you sure? (Y/n): ") {
        os.Exit(0)
    }
}
```

**Flag Binding:**

```go
cmd.Flags().BoolP("follow", "f", false, "Follow log output")
follow, _ := cmd.Flags().GetBool("follow")
```

### Testing Conventions

* Use standard `testing` package
* Table-driven tests where appropriate
* Test files located alongside implementation (`*_test.go`)
* Mock data in `test/` directory
* Target: 70%+ coverage on core logic
* Current coverage:
  * Config: 94.4%
  * Transformations: 100%
  * Utils: 16.9%

### Formatting and Style

* Standard Go formatting (`gofmt`)
* No emojis in code (only in output where appropriate, e.g., status icons)
* Clear, descriptive error messages
* Consistent command descriptions

## Completed Features (Phases 1-5)

### Phase 1: Foundation & Polish

* ✅ Test infrastructure (94.4% config coverage)
* ✅ Code cleanup and consistency

### Phase 2: User-Facing Improvements

* ✅ Enhanced list output (colors, status icons, table formatting)
* ✅ `--recreate` flag for create command
* ✅ `logs` command with follow, tail, timestamps

### Phase 3: Bulk Operations & Customization

* ✅ Bulk actions (`--all`, `--running`, `--stopped`)
* ✅ CLI customization flags (`--port`, `--volume`, `--env`, `--network`)

### Phase 4: Documentation

* ✅ Standalone `bin/gendocs.go` tool
* ✅ Auto-generated man pages and CLI reference
* ✅ `task docs` command
* ✅ Not included in shipped binary

### Phase 5: Docker Compose Support

* ✅ Unified interface for Docker and Compose services
* ✅ Compose CLI integration (`docker compose`)
* ✅ Path resolution (relative/absolute, file/directory)
* ✅ Folder name validation
* ✅ `--compose` filter for list command
* ✅ `--build` flag for compose services
* ✅ Service name used as compose project name

## Current State

### Version

**0.0.10** (defined in `cmd/root.go` via `GitTag`)

### Container Naming

* **Docker services:** `{prefix}-{serviceName}` (default prefix: `sda-`)
* **Compose services:** Managed by compose using service name as project name

### Supported Services (Default Config)

* PostgreSQL, MySQL, MSSQL, MariaDB, MongoDB, SurrealDB
* Redis, Memcached, Elasticsearch, RabbitMQ, Kafka
* Compose services (user-defined)

### Service Lifecycle

**Docker Service:**

1. `create` → creates and starts container
2. `stop` → stops container (preserves state)
3. `start` → restarts stopped container
4. `remove` → removes container (optionally volumes)

**Compose Service:**

1. `create` → runs `docker compose up` (creates and starts)
2. `stop` → runs `docker compose stop` (preserves state)
3. `start` → runs `docker compose start` (restarts)
4. `remove` → runs `docker compose down` (removes stack)

## Pending Work

### Phase 6: Distribution & Installation (NEXT)

* Package managers: Scoop, Chocolatey, WinGet, Homebrew, apt, yum, AUR, snap
* Installers and automation
* See `_dev/NEXT.md` for detailed implementation plan

### Phase 7: Service-Specific Fixes

* Elasticsearch CLI connect command fix

## Important Implementation Details

### Compose Path Resolution

**Relative Paths:**

```yaml
compose: ./myapp/docker-compose.yml
# Resolves to: ~/.config/sda/myapp/docker-compose.yml
```

**Absolute Paths:**

```yaml
compose: /home/user/projects/myapp/docker-compose.yml
# Used as-is
```

**Directory Paths:**

```yaml
compose: ./myapp
# Searches for: docker-compose.yaml or docker-compose.yml in ~/.config/sda/myapp/
```

**Validation:**

* Folder containing compose file MUST match service name
* Example: Service `myapp` requires folder `myapp/docker-compose.yml` ✓
* Example: Service `myapp` with `different/docker-compose.yml` ✗ (fails)

### Compose vs Docker Detection

The application automatically detects service type and routes to appropriate handler:

```go
// In internal/config/config.go
func (s *Service) IsComposeService() bool {
    return s.Compose != ""
}

// In cmd/*.go
service := config.CONFIG.GetServiceByName(name)
if service != nil && service.IsComposeService() {
    // Use compose operations
} else {
    // Use docker operations
}
```

### Flag Behavior

**Flags that work with both:**

* `--recreate` (create command)
* `--volumes` (remove command)
* `--yes` (all commands)

**Compose-only flags:**

* `--build` (create command) - Shows warning if used with Docker service

**Docker-only flags:**

* `--port`, `--volume`, `--env`, `--network`, `--password`, `--version` (create command)
* Ignored for compose services

### Output Format

**Commands support two output modes:**

1. **Table** (default) - Human-readable with colors and icons
2. **JSON** (via `--json`) - Machine-readable for scripting

**Status Icons:**

* ✓ (green) - Running
* ✗ (red) - Stopped
* ● (yellow) - Created
* 📦 (yellow) - Compose service

## Development Workflow

### Adding a New Command

1. Create `cmd/{command}.go`
2. Define cobra.Command with Use, Short, Long, Run
3. Add to `rootCmd` in `init()`
4. Implement service type detection
5. Call appropriate Docker or Compose method
6. Add tests if modifying core logic
7. Regenerate docs: `task docs`

### Adding a New Docker Operation

1. Add method to `internal/docker/operations.go` or `internal/docker/compose.go`
2. Use receiver `*Api` for methods; read config from `d.cfg`, not `config.CONFIG`
3. Construct clients with `docker.New(cfg)` which returns `(*Api, error)`
4. Handle errors appropriately
5. Return meaningful error messages
6. Add tests in `internal/docker/*_test.go` — construct `&Api{cfg: &config.Config{...}}` rather than assigning the package global

### Adding a New Config Field

1. Update struct in `internal/config/config.go`
2. Add `mapstructure` tag
3. Update `cmd/defaultConfig.yaml` with default value
4. Add tests in `internal/config/config_test.go`
5. Update AGENTS.md (this file)

## Versioning and Releases

* **Version Location:** `cmd/root.go` - `GitTag` variable set via build flags
* **Version Management:** Manual git tags (e.g., `v0.0.10`)
* **Build Script:** `scripts/increment_version.ts` for automated tag creation
* **Task Command:** `task version:increment [patch|minor|major]`
* **Release Process:**
  1. Create git tag manually or via script
  2. GitHub Actions automatically builds and publishes release
  3. Binaries for all platforms uploaded to GitHub Releases

## Dependencies

### Runtime Dependencies

* `github.com/spf13/cobra` - CLI framework
* `github.com/spf13/viper` - Configuration management
* `github.com/erikgeiser/promptkit` - Interactive prompts

No Docker SDK dependency: `internal/docker` shells out to the `docker` binary via `os/exec`
instead (see `internal/docker/exec.go`). This was a deliberate simplification - sda is a thin
CLI wrapper, and pulling in `github.com/docker/docker` + `github.com/docker/compose/v5` dragged
in ~150 indirect packages (buildx, buildkit, containerd, OpenTelemetry, gRPC, ...) to make a
dozen API calls.

### Dev Dependencies

* `github.com/spf13/cobra/doc` - Documentation generation
* Standard testing packages

### Notes

* No testing framework dependencies (uses standard `testing` package)
* Compose operations require the `docker compose` CLI plugin to be installed
* All dependencies managed via `go.mod`

## File References

### Key Files to Know

* `cmd/root.go:38` - `GetRootCommand()` exports root for docs generation
* `internal/config/config.go:27` - `IsComposeService()` service type detection
* `internal/docker/compose.go` - `resolveComposePath()` path resolution logic, `composeArgs()` shared `docker compose -f ... -p ...` argv, `ComposeUp()` etc.
* `internal/docker/create.go` - `buildCreateArgs()` translates a `config.Service` into `docker create` argv
* `internal/docker/exec.go` - `capture`/`run`/`runInteractive`, the only `os/exec` call sites for the `docker` binary
* `cmd/create.go:33` - Compose service handling in create command
* `cmd/start.go:113` - Compose service handling in start command

### Important Patterns

**Container Prefix:**

```go
// In internal/docker/operations.go — Api method, prefix comes from d.cfg
func (d *Api) containerName(name string) string {
    return fmt.Sprintf("%s-%s", d.cfg.Prefix, name)
}
```

**Service Lookup:**

```go
service := config.CONFIG.GetServiceByName(name)
if service == nil {
    utils.ErrorAndExit(fmt.Sprintf("Service '%s' not found", name))
}
```

**Compose Project Name:**

```go
// Service name is used as the compose project name (docker compose -p <name>)
args, err := d.composeArgs(service)
// -> ["compose", "-f", composePath, "-p", service.Name]
```

## Notes for Future Development

* Compose integration shells out to the `docker compose` CLI plugin - it inherits Docker's own
  progress rendering and log formatting for free, and needs no SDK dependency
* All commands route through unified interface - no separate compose commands
* Service type is transparent to users - detected automatically
* Path resolution is deterministic - relative to config directory
* Folder naming is enforced - prevents confusion and errors
* Documentation is auto-generated - keep command descriptions up to date
* Tests must pass before commits - CI enforces this
* Manual versioning is preferred - gives control over release timing
