# DII — Docker Image Inspector

Inspect Docker/OCI image metadata without pulling image layers.

**[中文](README_zh.md)**

---

## Installation

```bash
curl -fsSL https://raw.githubusercontent.com/DDDPS/dii/main/install.sh | bash
```

Or download a pre-built binary from [Releases](https://github.com/DDDPS/dii/releases), or build from source (requires Go 1.26+):

```bash
git clone https://github.com/DDDPS/dii.git && cd dii
go build -o dii main.go
```

---

## Usage

```text
dii [flags] IMAGE

  -o, --output string   Output format: md | json | yaml | compose (default: yaml)
  -d, --daemon          Inspect via local Docker daemon
  -r, --registry        Inspect via remote registry
  -t, --tar             Inspect a docker save archive (.tar / .tar.gz / .tgz)
  -w, --write string    Write output to a file (0644)
```

| Task | Command | Example |
|------|---------|---------|
| Auto-detect source | `dii ubuntu:latest` | [example](examples/dii_ubuntu_latest.yaml) |
| Local Docker daemon | `dii -d mysql:8.0` | [example](examples/dii_daemon_mysql_8.0.yaml) |
| Remote registry | `dii -r nginx:alpine` | [example](examples/dii_registry_nginx_alpine.yaml) |
| Tar archive | `dii -t image.tar.gz` | [example](examples/dii_tar_my_image.md) |
| JSON output | `dii redis:alpine -o json` | [example](examples/dii_redis_alpine.json) |
| Markdown output | `dii redis:alpine -o md` | [example](examples/dii_redis_alpine.md) |
| Generate Compose | `dii -r ghcr.io/gohugoio/hugo -o compose` | [example](examples/dii_hugo_compose.yaml) |
| Write to file | `dii nginx -o compose -w docker-compose.yaml` | [example](examples/docker-compose.yaml) |

When no source flag is given, `dii` tries in order: **Tar → Local Daemon → Remote Registry**.

> Only one source flag (`-d`, `-r`, `-t`) may be used at a time.

---

## Environment Variables

**`DII_MIRROR`** — Registry mirrors for remote inspection, useful on machines without Docker.

```bash
export DII_MIRROR="mirror1.example.com,mirror2.example.com"
```

Mirror resolution order: `DII_MIRROR` → `/etc/docker/daemon.json` → `docker.io`

## Issues
If you encounter any bugs, have questions, or feature requests, feel free to open an issue on [GitHub Issues](https://github.com/DDDPS/dii/issues).

---

[MIT](LICENSE)