# Gibson Tools Development Guide

This guide covers how to develop, build, and install Gibson tools. Tools are wrappers around external security binaries that expose them via gRPC for use by Gibson agents.

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Component Manifest (component.yaml)](#component-manifest-componentyaml)
- [Build System](#build-system)
- [Tool Installation](#tool-installation)
- [External Dependencies](#external-dependencies)
- [Creating a New Tool](#creating-a-new-tool)
- [Tool Registry](#tool-registry)
- [Security Policy](#security-policy)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

---

## Architecture Overview

Gibson tools follow a **wrapper pattern**:

```
┌─────────────────┐     gRPC      ┌──────────────────┐     exec      ┌─────────────────┐
│  Gibson Agent   │ ───────────▶ │   Tool Wrapper   │ ───────────▶ │ External Binary │
│                 │ ◀─────────── │   (Go binary)    │ ◀─────────── │  (nmap, etc.)   │
└─────────────────┘   JSON I/O   └──────────────────┘   stdout/XML  └─────────────────┘
```

**Key Points:**
- Tool wrappers are Go binaries that implement the Gibson Tool interface
- They execute external security tools (nmap, sqlmap, etc.) as subprocesses
- Input/output is converted between JSON (gRPC) and tool-native formats (XML, text, JSON lines)
- Health checks verify that external binaries are available

---

## Component Manifest (component.yaml)

Every tool must have a `component.yaml` manifest file in its directory. This file tells Gibson how to build, run, and validate the tool.

### Manifest Structure

```yaml
kind: tool                    # Component type: tool, agent, or plugin
name: nmap                    # Unique tool name (must match directory name)
version: 1.0.0               # Semantic version
description: Network scanner for port discovery and service detection
author: Gibson Security Team
license: MIT
repository: https://github.com/zero-day-ai/gibson-tools-official

build:
  command: go build -o nmap . # Build command (executed in workdir)
  artifacts:                  # Expected build outputs
    - nmap
  workdir: .                  # Working directory for build (relative to manifest)
  env:                        # Environment variables for build
    CGO_ENABLED: "0"

runtime:
  type: go                    # Runtime type: go, python, docker, binary
  entrypoint: ./nmap          # Executable path (relative to component dir)
  port: 0                     # gRPC port (0 = dynamic assignment)
  args: []                    # Additional command-line arguments
  env: {}                     # Runtime environment variables
  workdir: .                  # Runtime working directory

dependencies:
  gibson: ">=1.0.0"           # Minimum Gibson version
  system:                     # Required system binaries
    - nmap                    # External tool that must be installed
  components: []              # Other Gibson components required
  env:                        # Required environment variables
    API_KEY: "Description"    # Key: Description of required value
```

### Field Reference

#### Top-Level Fields

| Field | Required | Description |
|-------|----------|-------------|
| `kind` | Yes | Component type: `tool`, `agent`, or `plugin` |
| `name` | Yes | Unique identifier (alphanumeric, dash, underscore) |
| `version` | Yes | Semantic version (e.g., `1.0.0`) |
| `description` | No | Brief description of the tool |
| `author` | No | Author name or organization |
| `license` | No | License identifier (MIT, Apache-2.0, etc.) |
| `repository` | No | Source repository URL |

#### Build Configuration

| Field | Required | Description |
|-------|----------|-------------|
| `build.command` | No | Build command (default: `make build`) |
| `build.artifacts` | No | Expected output files |
| `build.workdir` | No | Build working directory |
| `build.env` | No | Build environment variables |
| `build.dockerfile` | No | Dockerfile path for container builds |

#### Runtime Configuration

| Field | Required | Description |
|-------|----------|-------------|
| `runtime.type` | Yes | Runtime type: `go`, `python`, `docker`, `binary` |
| `runtime.entrypoint` | Yes | Executable path or command |
| `runtime.port` | No | gRPC port (0 = dynamic) |
| `runtime.args` | No | Command-line arguments |
| `runtime.env` | No | Runtime environment variables |
| `runtime.image` | No | Docker image (for docker runtime) |

#### Dependencies

| Field | Required | Description |
|-------|----------|-------------|
| `dependencies.gibson` | No | Minimum Gibson version requirement |
| `dependencies.system` | No | Required system binaries |
| `dependencies.components` | No | Required Gibson components |
| `dependencies.env` | No | Required environment variables |

---

## Build System

### Using Make

The Gibson tools repository includes a Makefile for building all tools:

```bash
# Build all tools
make build

# Build specific phase
make build-recon          # Reconnaissance tools
make build-discovery      # Discovery tools
make build-initial-access # Initial access tools

# Run tests
make test                 # Unit tests
make integration-test     # Integration tests (requires external binaries)

# Clean build artifacts
make clean

# Show help
make help
```

### Using build.sh

Alternatively, use the shell script:

```bash
./build.sh              # Build all tools
./build.sh clean        # Clean artifacts
./build.sh test         # Run tests
./build.sh -v all       # Verbose build
```

### Build Output

Binaries are output to the `bin/` directory:

```
bin/
├── nmap
├── subfinder
├── httpx
├── sqlmap
└── ...
```

---

## Tool Installation

### Installing from Repository

Use the Gibson CLI to install tools from a git repository:

```bash
# Install from dedicated repository
gibson tool install https://github.com/user/my-tool

# Install from mono-repo subdirectory (use # fragment)
gibson tool install https://github.com/zero-day-ai/gibson-tools-official#discovery/nmap

# Install using SSH URL with subdirectory
gibson tool install git@github.com:zero-day-ai/gibson-tools-official.git#discovery/nmap

# Install with specific branch
gibson tool install https://github.com/user/my-tool --branch main

# Install with specific tag
gibson tool install https://github.com/user/my-tool --tag v1.0.0

# Force reinstall
gibson tool install https://github.com/user/my-tool --force

# Bulk install all tools from mono-repo
gibson tool install-all https://github.com/zero-day-ai/gibson-tools-official
```

### Installation Process

When you run `gibson tool install <repo-url>`:

1. **Parse URL**: Extract repository URL and optional subdirectory (from `#` fragment)
2. **Clone Repository**: Clone to temporary directory
3. **Locate Manifest**: Look for `component.yaml` in root (or subdirectory if specified with `#`)
4. **Validate Manifest**: Parse and validate manifest structure
5. **Check Dependencies**: Verify system dependencies are available
6. **Build Component**: Execute build command (default: `make build`)
7. **Install**: Move to `~/.gibson/tools/<name>/`
8. **Register**: Add to component registry

### Installation Directory Structure

```
~/.gibson/
├── tools/
│   ├── nmap/
│   │   ├── component.yaml
│   │   ├── go.mod
│   │   ├── main.go
│   │   ├── tool.go
│   │   └── nmap          # Built binary
│   ├── subfinder/
│   └── ...
├── agents/
├── plugins/
└── config.yaml
```

### Managing Installed Tools

```bash
# List installed tools
gibson tool list

# Get tool info
gibson tool info nmap

# Update a tool
gibson tool update nmap

# Update all tools
gibson tool update --all

# Uninstall a tool
gibson tool uninstall nmap

# Check tool health
gibson tool health nmap
```

---

## External Dependencies

Gibson tools wrap external security binaries. These must be installed separately.

### System Requirements by Tool

#### Reconnaissance Tools

| Tool | Binary | Installation |
|------|--------|--------------|
| subfinder | `subfinder` | `go install github.com/projectdiscovery/subfinder/v2/cmd/subfinder@latest` |
| httpx | `httpx` | `go install github.com/projectdiscovery/httpx/cmd/httpx@latest` |
| nuclei | `nuclei` | `go install github.com/projectdiscovery/nuclei/v3/cmd/nuclei@latest` |
| amass | `amass` | `go install github.com/owasp-amass/amass/v4/...@master` |
| theharvester | `theHarvester` | `pip install theHarvester` |
| recon-ng | `recon-ng` | `pip install recon-ng` |
| shodan | `shodan` | `pip install shodan` (requires API key) |
| spiderfoot | `spiderfoot` | `pip install spiderfoot` |
| playwright | `chromium` | `npx playwright install` |

#### Discovery Tools

| Tool | Binary | Installation |
|------|--------|--------------|
| nmap | `nmap` | `apt install nmap` |
| masscan | `masscan` | `apt install masscan` |
| crackmapexec | `crackmapexec` | `pip install crackmapexec` or `pipx install crackmapexec` |
| bloodhound | `bloodhound-python` | `pip install bloodhound` |

#### Initial Access Tools

| Tool | Binary | Installation |
|------|--------|--------------|
| sqlmap | `sqlmap` | `apt install sqlmap` |
| gobuster | `gobuster` | `apt install gobuster` or `go install github.com/OJ/gobuster/v3@latest` |
| hydra | `hydra` | `apt install hydra` |
| metasploit | `msfconsole` | Install Metasploit Framework |

#### Execution Tools

| Tool | Binary | Installation |
|------|--------|--------------|
| evil-winrm | `evil-winrm` | `gem install evil-winrm` |
| impacket | `impacket-*` | `pip install impacket` |

#### Privilege Escalation Tools

| Tool | Binary | Installation |
|------|--------|--------------|
| hashcat | `hashcat` | `apt install hashcat` |
| john | `john` | `apt install john` |
| linpeas | `linpeas.sh` | Download from PEASS-ng releases |
| winpeas | `winpeas.exe` | Download from PEASS-ng releases |

#### Credential Access Tools

| Tool | Binary | Installation |
|------|--------|--------------|
| responder | `Responder.py` | `git clone https://github.com/lgandx/Responder` |
| secretsdump | `impacket-secretsdump` | `pip install impacket` |

#### Other Tools

| Tool | Binary | Installation |
|------|--------|--------------|
| chisel | `chisel` | `go install github.com/jpillora/chisel@latest` |
| msfvenom | `msfvenom` | Install Metasploit Framework |
| proxychains | `proxychains4` | `apt install proxychains4` |
| xfreerdp | `xfreerdp` | `apt install freerdp2-x11` |
| tshark | `tshark` | `apt install tshark` |
| sliver | `sliver-client` | Install Sliver C2 |
| rclone | `rclone` | `apt install rclone` |
| slowhttptest | `slowhttptest` | `apt install slowhttptest` |

### Installing All Dependencies (Debian/Ubuntu)

```bash
# System packages
sudo apt update
sudo apt install -y \
    nmap masscan tshark hydra hashcat john \
    gobuster proxychains4 freerdp2-x11 \
    sqlmap slowhttptest rclone

# Go tools
go install github.com/projectdiscovery/subfinder/v2/cmd/subfinder@latest
go install github.com/projectdiscovery/httpx/cmd/httpx@latest
go install github.com/projectdiscovery/nuclei/v3/cmd/nuclei@latest
go install github.com/owasp-amass/amass/v4/...@master
go install github.com/OJ/gobuster/v3@latest
go install github.com/jpillora/chisel@latest

# Python tools
pip install impacket bloodhound crackmapexec theHarvester recon-ng shodan spiderfoot

# Ruby tools
gem install evil-winrm

# Playwright browsers
npx playwright install
```

---

## Creating a New Tool

### 1. Create Directory Structure

```bash
mkdir -p mytool
cd mytool
```

### 2. Create go.mod

```go
module github.com/zero-day-ai/gibson-tools-official/category/mytool

go 1.24

require (
    github.com/zero-day-ai/gibson-tools-official/pkg v0.0.0
    github.com/zero-day-ai/sdk v0.0.0
)

replace (
    github.com/zero-day-ai/gibson-tools-official/pkg => ../../pkg
    github.com/zero-day-ai/sdk => ../../../sdk
)
```

### 3. Create schema.go

```go
package main

import "github.com/zero-day-ai/sdk/schema"

func InputSchema() *schema.Schema {
    return schema.NewObject().
        Property("target", schema.NewString().
            Description("Target to scan").
            Required()).
        Property("options", schema.NewObject().
            Description("Additional options")).
        Build()
}

func OutputSchema() *schema.Schema {
    return schema.NewObject().
        Property("results", schema.NewArray(schema.NewObject())).
        Property("scan_time_ms", schema.NewInteger()).
        Build()
}
```

### 4. Create tool.go

```go
package main

import (
    "context"
    "time"

    "github.com/zero-day-ai/gibson-tools-official/pkg/executor"
    "github.com/zero-day-ai/sdk/tool"
    "github.com/zero-day-ai/sdk/types"
)

const (
    ToolName    = "mytool"
    ToolVersion = "1.0.0"
    BinaryName  = "mytool-binary"
)

type ToolImpl struct{}

func NewTool() tool.Tool {
    cfg := tool.NewConfig().
        SetName(ToolName).
        SetVersion(ToolVersion).
        SetDescription("My tool description").
        SetTags([]string{"category", "T1234"}).
        SetInputSchema(InputSchema()).
        SetOutputSchema(OutputSchema()).
        SetExecuteFunc((&ToolImpl{}).Execute)

    t, _ := tool.New(cfg)
    return &toolWithHealth{Tool: t, impl: &ToolImpl{}}
}

type toolWithHealth struct {
    tool.Tool
    impl *ToolImpl
}

func (t *toolWithHealth) Health(ctx context.Context) types.HealthStatus {
    return t.impl.Health(ctx)
}

func (t *ToolImpl) Execute(ctx context.Context, input map[string]any) (map[string]any, error) {
    start := time.Now()

    // Build command arguments
    args := []string{"--option", input["target"].(string)}

    // Execute external binary
    result, err := executor.Execute(ctx, executor.Config{
        Command: BinaryName,
        Args:    args,
        Timeout: 5 * time.Minute,
    })
    if err != nil {
        return nil, err
    }

    // Parse output and return
    return map[string]any{
        "results":      parseOutput(result.Stdout),
        "scan_time_ms": time.Since(start).Milliseconds(),
    }, nil
}

func (t *ToolImpl) Health(ctx context.Context) types.HealthStatus {
    if !executor.BinaryExists(BinaryName) {
        return types.NewUnhealthyStatus("binary not found", nil)
    }
    return types.NewHealthyStatus("ready")
}
```

### 5. Create main.go

```go
package main

import (
    "log"
    "github.com/zero-day-ai/sdk/serve"
)

func main() {
    if err := serve.Tool(NewTool()); err != nil {
        log.Fatal(err)
    }
}
```

### 6. Create component.yaml

```yaml
kind: tool
name: mytool
version: 1.0.0
description: My custom security tool
author: Your Name
license: MIT
repository: https://github.com/your-org/your-repo

build:
  command: go build -o mytool .
  artifacts:
    - mytool
  workdir: .

runtime:
  type: go
  entrypoint: ./mytool
  port: 0

dependencies:
  gibson: ">=1.0.0"
  system:
    - mytool-binary
  env: {}
```

### 7. Build and Test

```bash
# Build
go build -o mytool .

# Test health check
./mytool &
gibson tool health mytool

# Run integration tests
go test -v -tags=integration .
```

---

## Tool Registry

The Gibson component registry tracks installed tools and their status.

### Registry Location

```
~/.gibson/registry.json
```

### Registry Operations

```go
// Get registry
registry := component.NewFileRegistry("~/.gibson")

// Register a component
registry.Register(component)

// List components
tools := registry.List(component.ComponentKindTool)

// Get specific component
tool := registry.Get(component.ComponentKindTool, "nmap")

// Unregister
registry.Unregister(component.ComponentKindTool, "nmap")
```

### Component States

| State | Description |
|-------|-------------|
| `available` | Installed and ready to use |
| `running` | Currently executing |
| `stopped` | Manually stopped |
| `error` | Failed to start or crashed |
| `updating` | Being updated |

---

## Security Policy

### No Binaries in Repository

**CRITICAL**: The Gibson Tools repository contains **source code only**. No pre-compiled binaries are permitted for security reasons.

#### Why Source Code Only?

1. **Supply Chain Security**: Pre-compiled binaries cannot be audited and may contain malicious code, backdoors, or vulnerabilities not present in the source code. In offensive security tooling, this risk is especially severe.

2. **Transparency**: All code must be reviewable. Security researchers and users must be able to verify exactly what they are running before deploying in sensitive environments.

3. **Reproducible Builds**: Building from source ensures the binary matches the source code and hasn't been tampered with during distribution.

4. **Trust Verification**: Offensive security tools operate in high-trust environments. Source code allows security teams to audit tool behavior before deployment.

#### What This Means

**DO NOT commit:**
- Compiled binaries or executables
- Files in `bin/` directories
- `.exe`, `.dll`, `.so`, `.dylib`, or other binary formats
- Pre-built tool wrappers

**DO commit:**
- Source code (`.go`, `.py`, `.yaml`, `.json`, etc.)
- Build scripts and Makefiles
- Documentation and tests

#### Build Artifacts

All tool wrappers must be built locally from source:

```bash
# Build all tools
make build

# Build specific tool
cd discovery/nmap && go build -o nmap .
```

The `bin/` directory is for local builds only and is excluded via `.gitignore`. Built artifacts should never be committed.

#### Contributor Requirements

Before submitting a PR:
1. Ensure no binaries are included in the commit
2. Verify `git status` shows only source files
3. Test that the tool builds from source
4. Remove any accidentally committed binaries

---

## Best Practices

1. **Always validate input** - Use schema validation before executing external commands
2. **Handle timeouts** - Set reasonable timeouts for external commands
3. **Parse output carefully** - External tools may change output format between versions
4. **Check health on startup** - Verify external dependencies before accepting requests
5. **Use structured logging** - Log command execution and errors for debugging
6. **Don't hardcode paths** - Use PATH lookup for external binaries
7. **Clean up temp files** - Remove temporary files created during execution
8. **Tag with MITRE ATT&CK** - Add relevant technique IDs to tool tags

---

## Troubleshooting

### Tool not found after install

```bash
# Check if installed
gibson tool list

# Check installation path
ls -la ~/.gibson/tools/

# Reinstall
gibson tool install <url> --force
```

### External binary not found

```bash
# Check if binary is in PATH
which nmap

# Check tool health
gibson tool health nmap
```

### Build fails

```bash
# Check Go version
go version

# Clean and rebuild
make clean && make build

# Check dependencies
go mod tidy
```

### Permission denied

```bash
# Some tools require root/capabilities
sudo setcap cap_net_raw+ep $(which nmap)

# Or run as root (not recommended)
sudo gibson tool run nmap
```
