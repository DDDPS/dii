# Image Metadata

## Basic Information
- **ID**: `sha256:1d40e3eb3bf4f138de1d67193f2aa5309fcaf343eb5ffadbf5e9439de1eb1ebb`
- **RepoTags**:
  - `nginx:alpine`
- **OS**: `linux`
- **Architecture**: `amd64`
- **Size**: `26.1 MB`
- **Created**: `2026-07-15T23:57:33Z`

## Config
- **WorkingDir**: `/`
- **StopSignal**: `SIGQUIT`
- **ExposedPorts**:
  - `80/tcp`
- **Env**:
  - `PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin`
  - `NGINX_VERSION=1.31.3`
  - `PKG_RELEASE=1`
  - `DYNPKG_RELEASE=1`
  - `NJS_VERSION=1.0.0`
  - `NJS_RELEASE=1`
  - `ACME_VERSION=0.4.1`
- **Entrypoint**:
  - `/docker-entrypoint.sh`
- **Cmd**:
  - `nginx`
  - `-g`
  - `daemon off;`

## Labels
```json
{
  "maintainer": "NGINX Docker Maintainers \u003cdocker-maint@nginx.com\u003e"
}
```
