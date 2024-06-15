set windows-shell := ["pwsh.exe", "-NoLogo", "-Command"]

app_name := if os_family() == "windows" { "sda.exe" } else { "sda" }

[private]
default:
  @just --list --justfile {{justfile()}}

# Build application
@build:
        echo "Building {{ app_name }}..."
        go build -o target/{{ app_name }} main.go

# Build for all platforms
[unix]
@build-all:
        echo "Building for 'windows/amd64'..."
        env GOOS=windows GOARCH=amd64 go build -o target/{{ app_name }}-windows-x64.exe main.go
        echo "Building for 'windows/arm64'..."
        env GOOS=windows GOARCH=arm64 go build -o target/{{ app_name }}-windows-arm64.exe main.go
        echo "Building for 'linux/amd64'..."
        env GOOS=linux GOARCH=amd64 go build -o target/{{ app_name }}-linux-x64 main.go
        echo "Building for 'linux/arm64'..."
        env GOOS=linux GOARCH=arm64 go build -o target/{{ app_name }}-linux-arm64 main.go

# Build for all platforms
[windows]
@build-all:
        echo "Building for 'windows/amd64'..."
        $Env:GOOS="windows"; $Env:GOARCH="amd64"; go build -o target/{{ app_name }}-windows-x64.exe main.go
        echo "Building for 'windows/arm64'..."
        $Env:GOOS="windows"; $Env:GOARCH="arm64"; go build -o target/{{ app_name }}-windows-arm64.exe main.go
        echo "Building for 'linux/amd64'..."
        $Env:GOOS="linux"; $Env:GOARCH="amd64"; go build -o target/{{ app_name }}-linux-x64 main.go
        echo "Building for 'linux/arm64'..."
        $Env:GOOS="linux"; $Env:GOARCH="arm64"; go build -o target/{{ app_name }}-linux-arm64 main.go

# Test the application
@test:
        echo "Testing..."
        go test -v ./...

# Clean the binary
[unix]
@clean:
        echo "Cleaning..."
        rm -rf target
        go clean

# Clean the binary
[windows]
@clean:
        echo "Cleaning..."
        if (Test-Path target) { rm -r -fo target }
        go clean
