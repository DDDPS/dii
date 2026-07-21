# DII — Docker Image Inspector

一个用于提取 Docker/OCI 镜像元数据的命令行工具，支持多种数据源与输出格式。

**[English](README.md)**

---

## 特性

数据来源：

- **本地 Docker Daemon**（`-d`）：通过本地Docker守护进程查询本地已有镜像
- **远程 Registry**（`-r`）：仅获取远程镜像的配置文件，不下载镜像层
- **Tar包**（`-t`）：解析 `docker save` 导出的 `.tar`、`.tar.gz`、`.tgz`，流式解压，不占用额外磁盘空间

不指定来源时，自动按 `Tar包 → 本地 Daemon → 远程 Registry` 顺序尝试。

支持以下格式输出（`-o`）：

- `md`（默认）/ `json` / `yaml`：以不同格式输出相同的镜像元数据
- `compose`：根据镜像配置生成参数化的 `docker-compose.yaml`（如 `${APP_PORT:-8080}`）

---

## 安装

```bash
curl -fsSL https://raw.githubusercontent.com/DDDPS/dii/main/install.sh | bash
```


### 手动下载

[Releases](https://github.com/DDDPS/dii/releases) 下载二进制文件。

### 源码编译

**前提：** Go 1.26+

```bash
git clone https://github.com/DDDPS/dii.git && cd dii
go build -o dii main.go
```

---

## 快速开始

| 功能 | 命令 | 说明 | 输出示例 |
| :--- | :--- | :--- | :--- |
| **自动检测** | `dii ubuntu:latest` | 自动匹配数据源，默认输出 Markdown | [示例](examples/dii_ubuntu_latest.md) |
| **本地 Daemon** | `dii -d mysql:8.0` | 显式指定本地 Docker Daemon | [示例](examples/dii_daemon_mysql_8.0.md) |
| **远程 Registry** | `dii -r nginx:alpine` | 显式指定远程 Registry（不拉取镜像层） | [示例](examples/dii_registry_nginx_alpine.md) |
| **Tar 包** | `dii -t ./my-image.tar.gz` | 解析本地Tar归档文件 | [示例](examples/dii_tar_my_image.md) |
| **JSON / YAML** | `dii redis:alpine -o json` / `-o yaml` | 输出原始镜像配置 | [JSON](examples/dii_redis_alpine.json) / [YAML](examples/dii_redis_alpine.yaml) |
| **生成 Compose** | `dii postgres -o compose` | 生成 `docker-compose.yaml` | [示例](examples/dii_postgres_compose.yaml) |
| **写入文件** | `dii ghcr.io/gohugoio/hugo -o compose -w docker-compose.yaml` | 将结果写入指定文件（0644） | [示例](examples/docker-compose.yaml) |

---

## 参数

```
dii [flags] IMAGE

  -o, --output string   输出格式：md / json / yaml / compose（默认 md）
  -d, --daemon          通过本地 Docker Daemon 获取
  -r, --registry        通过远程 Registry 获取
  -t, --tar             从本地归档文件获取（.tar / .tar.gz / .tgz）
  -w, --write string    将结果写入文件（权限 0644）
```

---

[MIT License](LICENSE)
