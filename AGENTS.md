# Project Context: sda (Simple Docker Apps)

## Overview

`sda` is a Go-based CLI tool designed to simplify the process of spinning up development servers (databases, caches, etc.) using Docker. It acts as a wrapper around the Docker SDK, allowing users to define services in a configuration file and manage them with simple commands like `sda create`, `sda start`, and `sda connect`.

It is intended for local development to avoid repetitive `docker run` commands and manual volume/port configuration.

## Key Technologies

* **Language:** Go (1.25.0)
* **CLI Framework:** [Cobra](https://github.com/spf13/cobra)
* **Configuration:** [Viper](https://github.com/spf13/viper)
* **Container Runtime:** [Docker SDK for Go](https://github.com/docker/docker/client)
* **Task Runner:** [Task](https://taskfile.dev/)
* **Interactive Prompts:** [Promptkit](https://github.com/erikgeiser/promptkit)

## Architecture

* **Entry Point:** `main.go` calls `cmd.Execute()`.
* **CLI Commands:** Defined in `cmd/` (e.g., `create.go`, `list.go`, `start.go`, `stop.go`, `remove.go`, `show.go`, `connect.go`, `root.go`).
* **Configuration:** Managed in `internal/config/`. The core structs are `Config`, `Service`, and `Docker`, mapping to the `sda.yaml` file.
* **Docker Logic:** Encapsulated in `internal/docker/` with modular files:
  * `api.go` - Docker client initialization
  * `create.go` - Container creation logic
  * `operations.go` - List, Start, Stop, Remove, Connect, GetInfo operations
  * `network.go` - Network management
  * `transformations.go` - Helper transformations (ports, versions, etc.)
  * `types.go` - Type definitions (ServiceInfo struct)
* **Utilities:** `internal/utils/` contains helper functions for output, prompts, command execution, and custom flags.

## Build and Run

### Prerequisites

* Go 1.22 or higher
* Docker (running)
* `task` (optional, for convenience)

### Commands (using `task`)

The project uses a `Taskfile.yaml` for common tasks.

* **Build:** `task build` (outputs to `publish/sda` or `publish/sda.exe`)
* **Run:** `task run [command]` (runs the built binary)
* **Build and Run:** `task runb` (builds then runs)
* **Test:** `task test` (runs `go test -v ./...`)
* **Build All Platforms:** `task build-all` (cross-compiles for Windows/Linux/Darwin AMD64/ARM64)
* **Clean:** `task clean`

### Commands (standard Go)

* **Build:** `go build -o publish/sda main.go`
* **Run:** `go run main.go [command]`

## Configuration

The application uses a configuration file typically located at `$HOME/.config/sda/sda.yaml`.

* On first run, if the config doesn't exist, a default one (embedded in `cmd/defaultConfig.yaml`) is created.
* The config defines default settings (network, password) and the list of supported services (images, ports, volumes).

## Development Conventions

* **Project Structure:** Follows standard Go CLI layout: `cmd/` for commands, `internal/` for private application code.
* **Formatting:** Standard Go formatting (`gofmt`).
* **Error Handling:** Custom `utils.ErrorAndExit` is often used for CLI-level fatal errors.
* **Dependencies:** Managed via `go.mod`.
* **Output Format:** Commands support JSON output via `--json` flag (using Viper binding).
* **Container Naming:** Services use a configurable prefix (default: `sda-`) for container naming.
* **Version:** Current version is 0.0.10 (defined in `cmd/root.go`).
