# Image Metadata

## Basic Information
- **ID**: `sha256:c66c7a7a13c9cb54f138de1d67193f2aa5309fcaf343eb5ffadbf5e9439de1eb1ebb`
- **RepoTags**:
  - `mysql:8.0`
- **OS**: `linux`
- **Architecture**: `amd64`
- **Size**: `412.3 MB`
- **Created**: `2026-07-10T12:00:00Z`

## Config
- **ExposedPorts**:
  - `3306/tcp`
  - `33060/tcp`
- **Env**:
  - `PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin`
  - `GOSU_VERSION=1.17`
  - `MYSQL_VERSION=8.0.40`
- **Entrypoint**:
  - `docker-entrypoint.sh`
- **Cmd**:
  - `mysqld`
- **Volumes**:
  - `/var/lib/mysql`
