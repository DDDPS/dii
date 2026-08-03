# DII — Docker Image Inspector

无需拉取镜像，即可查看 Docker/OCI 镜像元数据。

**[English](README.md)**

---

## 安装

```bash
curl -fsSL https://raw.githubusercontent.com/DDDPS/dii/main/install.sh | bash
```

从 [Releases](https://github.com/DDDPS/dii/releases) 下载预编译二进制文件，也可从源码构建（需要 Go 1.26+）：

```bash
git clone https://github.com/DDDPS/dii.git && cd dii
go build -o dii main.go
```

---

## 使用

```text
dii [flags] IMAGE

  -o, --output string   输出格式：md | json | yaml | compose（默认：yaml）
  -d, --daemon          通过本地 Docker 守护进程获取
  -r, --registry        通过远程镜像仓库获取
  -t, --tar             从 docker save 归档文件获取（.tar / .tar.gz / .tgz）
  -w, --write string    将输出写入文件（0644）
```

| 功能 | 命令 | 示例 |
|------|------|------|
| 自动检测数据源 | `dii ubuntu:latest` | [example](examples/dii_ubuntu_latest.yaml) |
| 本地 Docker 守护进程 | `dii -d mysql:8.0` | [example](examples/dii_daemon_mysql_8.0.yaml) |
| 远程镜像仓库 | `dii -r nginx:alpine` | [example](examples/dii_registry_nginx_alpine.yaml) |
| Tar 归档文件 | `dii -t image.tar.gz` | [example](examples/dii_tar_my_image.md) |
| JSON 输出 | `dii redis:alpine -o json` | [example](examples/dii_redis_alpine.json) |
| Markdown 输出 | `dii redis:alpine -o md` | [example](examples/dii_redis_alpine.md) |
| 生成 Compose | `dii -r ghcr.io/gohugoio/hugo -o compose` | [example](examples/dii_hugo_compose.yaml) |
| 写入文件 | `dii nginx -o compose -w docker-compose.yaml` | [example](examples/docker-compose.yaml) |

不指定来源时，自动按 **Tar → 本地 Daemon → 远程 Registry** 顺序尝试。

> 数据源参数（`-d`、`-r`、`-t`）同时只能使用一个。

---

## 环境变量

**`DII_MIRROR`** — 指定远程拉取时的加速源，在没有安装 Docker 的机器上同样有效。

```bash
export DII_MIRROR="mirror1.example.com,mirror2.example.com"
```

解析优先级：`DII_MIRROR` → `/etc/docker/daemon.json` → `docker.io`

---

[MIT](LICENSE)
