# DII — Docker Image Inspector

A command-line tool for extracting Docker/OCI image metadata, with support for multiple data sources and output formats.

**[中文](README_zh.md)**

---

## Features

Data sources:

- **Local Docker Daemon** (`-d`): Query locally available images via the local Docker daemon
- **Remote Registry** (`-r`): Fetch only the image config file from a remote registry, without downloading image layers
- **Tar archive** (`-t`): Parse archives exported by `docker save` (`.tar`, `.tar.gz`, `.tgz`), with streaming decompression and no extra disk usage

When no source is specified, the tool automatically tries in order: `Tar archive → Local Daemon → Remote Registry`.

Output formats (`-o`):

- `md` (default) / `json` / `yaml`: The same image metadata in different formats
- `compose`: Generate a parameterized `docker-compose.yaml` from the image config (e.g. `${APP_PORT:-8080}`)

---

## Installation

```bash
curl -fsSL https://raw.githubusercontent.com/DDDPS/dii/main/install.sh | bash
```

### Manual download

Download the binary from the [Releases](https://github.com/DDDPS/dii/releases) page.

### Build from source

**Prerequisite:** Go 1.26+

```bash
git clone https://github.com/DDDPS/dii.git && cd dii
go build -o dii main.go
```

---

## Quick Start

| Feature | Command | Description | Example Output |
| :--- | :--- | :--- | :--- |
| **Auto-detect** | `dii ubuntu:latest` | Auto-selects the data source; outputs Markdown by default | [Example](examples/dii_ubuntu_latest.md) |
| **Local Daemon** | `dii -d mysql:8.0` | Explicitly use the local Docker daemon | [Example](examples/dii_daemon_mysql_8.0.md) |
| **Remote Registry** | `dii -r nginx:alpine` | Explicitly use a remote registry (no layer download) | [Example](examples/dii_registry_nginx_alpine.md) |
| **Tar archive** | `dii -t ./my-image.tar.gz` | Parse a local tar archive | [Example](examples/dii_tar_my_image.md) |
| **JSON / YAML** | `dii redis:alpine -o json` / `-o yaml` | Output raw image config | [JSON](examples/dii_redis_alpine.json) / [YAML](examples/dii_redis_alpine.yaml) |
| **Generate Compose** | `dii postgres -o compose` | Generate a `docker-compose.yaml` | [Example](examples/dii_postgres_compose.yaml) |
| **Write to file** | `dii ghcr.io/gohugoio/hugo -o compose -w docker-compose.yaml` | Write output to a file (mode 0644) | [Example](examples/docker-compose.yaml) |

---

## Flags

```
dii [flags] IMAGE

  -o, --output string   Output format: md / json / yaml / compose (default: md)
  -d, --daemon          Fetch via local Docker daemon
  -r, --registry        Fetch via remote registry
  -t, --tar             Fetch from local tar archive (.tar / .tar.gz / .tgz)
  -w, --write string    Write output to a file (mode 0644)
```

---

[MIT License](LICENSE)
