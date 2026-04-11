# Docker 单镜像部署（Nginx + 前端静态资源 + 后端 API）

## 能不能「一个镜像」搞定前后端和 Nginx？

**可以。** 仓库提供 **`Dockerfile.all-in-one`**：在一个镜像里同时包含：

- **Nginx**：对外只暴露 **80**，提供静态前端、并把 **`/api` 反代**到同容器内的 Go 服务（监听 `8080`，不映射到宿主机）。
- **前端**：`npm run build` 后的静态文件，位于 `/usr/share/nginx/html`。
- **后端**：`bms-api` 可执行文件与 `server/_config`。

镜像构建时会把 `conf.yaml` 里的数据库地址从 `127.0.0.1` **替换为 `mysql`**，以便与 Compose 里的 MySQL 服务名一致（默认账号密码仍为 `root` / `123456`，与示例一致）。

## 不能「只靠一个容器」的部分：数据库

业务数据在 **MySQL** 里，仍需要 **第二个容器（或外部已有 MySQL）**。常见做法是：

- **官方 `mysql:8.0` 镜像** + 本项目的 **`database_schema.sql`** 做首次初始化。

因此：**最小可运行组合是 2 个容器**（`app` + `mysql`），而不是 1 个。

若坚持「只跑一个容器」，需要自备已可访问的 MySQL，并**挂载自定义 `conf.yaml`** 指向该库（见下文），一般不推荐给新手。

## 和原来的 `docker-compose.yml` 有什么区别？

| 项目 | 原 `docker-compose.yml` | `docker-compose.single.yml` |
|------|-------------------------|-----------------------------|
| 前端静态文件 | 挂载宿主机 `./frontend/build`，**未构建则页面为空** | 打进 `app` 镜像，**不依赖本机目录** |
| Nginx | 独立容器 + 挂载本机 `nginx/*.conf` | 在 `app` 镜像内，**无需单独 Nginx 容器** |
| 对外访问 | 80（Nginx）+ 8080（后端）等 | 一般只映射 **80** 即可 |

## 在本机从源码构建并启动（开发者）

仓库根目录执行：

```bash
docker compose -f docker-compose.single.yml up -d --build
```

浏览器访问：**http://localhost**

要求：与 `docker-compose.single.yml` **同目录**下存在 **`database_schema.sql`**（用于 MySQL 首次初始化）。

## 「不下载整个项目」时怎么用？

可以，但需要两样东西（无法做到完全零文件，至少要 **数据库初始化 SQL** 或等价数据）：

1. **应用镜像**：在能访问源码的机器上构建并推送到镜像仓库，例如：

   ```bash
   docker build -f Dockerfile.all-in-one -t <你的仓库>/labequip-bms:1.0 .
   docker push <你的仓库>/labequip-bms:1.0
   ```

2. **给使用方一个极简 `docker-compose.yml`**（或一条 `docker run`），其中 `app` 使用：

   ```yaml
   image: <你的仓库>/labequip-bms:1.0
   # 不要使用 build:，对方机器上无需源码
   ```

3. **`database_schema.sql`**：随发行包提供，或放在可下载地址，与 compose 放在同一目录，挂载到 MySQL 的 `/docker-entrypoint-initdb.d/`（见现有 `docker-compose.single.yml`）。

对方机器上只需：**Docker**、**compose 文件**、**database_schema.sql**、**拉取镜像**，无需 `git clone` 整个仓库。

## 自定义数据库密码或外部 MySQL

镜像内默认已将主机名写为 **`mysql`**、密码与仓库 `conf.yaml` 一致（`root` / `123456`）。若需修改：

- 在 **Compose** 里改 MySQL 的 `MYSQL_ROOT_PASSWORD`，并**挂载**你改好的 `server/_config/conf.yaml` 到容器的 `/root/server/_config/conf.yaml`，或
- 自行构建镜像时在 Dockerfile 中替换配置。

## 常见问题

### 端口

- **80**：Nginx（页面 + `/api` 反代）。
- **8080**：仅在容器**内部**给 Nginx 访问 Go，一般不必映射到宿主机。

### 与仅后端镜像 `Dockerfile` 的关系

根目录 **`Dockerfile`** 仍可用于「只跑 API、不配 Nginx」的场景；**`Dockerfile.all-in-one`** 面向「一个入口 80 的完整 Web」。

### 健康检查

`docker-compose.single.yml` 使用 MySQL 的 `healthcheck`，需 **Docker Compose v2** 一类较新版本。若环境过旧，可去掉 `depends_on` 里的 `condition: service_healthy`，改为 `depends_on: [mysql]`，并依赖入口脚本里的等待逻辑。

---

*若你优化了镜像分层或配置路径，请同步更新 `Dockerfile.all-in-one` 与本文。*
