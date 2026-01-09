# Project Context: sda (Simple Docker Apps)

## Overview

`sda` is a Go-based CLI tool designed to simplify the process of spinning up development servers (databases, caches, etc.) using Docker. It acts as a wrapper around the Docker SDK, allowing users to define services in a configuration file and manage them with simple commands like `sda create`, `sda start`, and `sda connect`.

It is intended for local development to avoid repetitive `docker run` commands and manual volume/port configuration.

## Key Technologies

* **Language:** Go (1.22+)
* **CLI Framework:** [Cobra](https://github.com/spf13/cobra)
* **Configuration:** [Viper](https://github.com/spf13/viper)
* **Container Runtime:** [Docker SDK for Go](https://github.com/docker/docker/client)
* **Task Runner:** [Just](https://github.com/casey/just)

## Architecture

* **Entry Point:** `main.go` calls `cmd.Execute()`.
* **CLI Commands:** Defined in `cmd/` (e.g., `create.go`, `list.go`, `root.go`).
* **Configuration:** Managed in `internal/config/`. The core structs are `Config`, `Service`, and `Docker`, mapping to the `sda.yaml` file.
* **Docker Logic:** Encapsulated in `internal/docker/`. Handles API interaction, container creation, and network management.
* **Utilities:** `internal/utils/` contains helper functions for output, prompts, and common logic.

## Build and Run

### Prerequisites

* Go 1.22 or higher
* Docker (running)
* `just` (optional, for convenience)

### Commands (using `just`)

The project uses a `justfile` for common tasks.

* **Build:** `just build` (outputs to `target/sda` or `target/sda.exe`)
* **Test:** `just test` (runs `go test ./...`)
* **Build All Platforms:** `just build-all` (cross-compiles for Linux/Windows AMD64/ARM64)
* **Clean:** `just clean`

### Commands (standard Go)

* **Build:** `go build -o target/sda main.go`
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
