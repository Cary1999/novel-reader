# novel-reader

本地小说阅读网站（MVP）：支持注册/登录、搜索、书籍详情、按章节阅读、上传 `.txt` 小说、后台内容管理。  
技术栈：后端 Go + Kratos，前端 Vite + React + TypeScript，数据库 MySQL，上传文件保存本地文件系统，支持 Docker/Compose 本地运行。

## 快速开始（推荐）

1. 准备环境

- Docker Desktop（或兼容的 Docker Engine + Compose）

2. 初始化环境变量

```bash
cp .env.example .env
```

3. 启动完整服务栈（前端 + 后端 + MySQL）

```bash
make dev
```

默认访问地址：

- 前端：`http://localhost:5173`
- 后端：`http://localhost:8000`（健康检查：`/api/health`）
- MySQL：`localhost:3306`

停止并清理容器（会停止服务，保留 MySQL volume）：

```bash
docker compose -f deploy/docker-compose.yml --env-file .env down --remove-orphans
```

## 混合启动（开发调试更顺手）

日常开发更推荐：Docker 只跑 MySQL，本地跑后端与前端，热更新和日志更直观。

1. 启动 MySQL

```bash
docker compose -f deploy/docker-compose.yml --env-file .env up -d mysql
```

2. 本地启动后端（示例）

```bash
cd backend
DATABASE_DSN="novel_reader:change-me-db-password@tcp(127.0.0.1:3306)/novel_reader?charset=utf8mb4&parseTime=True&loc=Local" \
HTTP_ADDR=":8000" \
go run ./cmd/server
```

3. 本地启动前端（Vite dev server）

```bash
cd frontend
npm run dev
```

前端会通过 Vite proxy 将 `/api/*` 代理到 `http://localhost:8000`（见 `frontend/vite.config.ts`）。

## 配置说明（.env）

项目通过 `.env` 控制 Compose 端口、MySQL 账号密码、后端 DSN、JWT 密钥、管理员账号等。

常用参数：

- `BACKEND_PORT`：后端端口（默认 `8000`）
- `FRONTEND_PORT`：前端端口（默认 `5173`）
- `MYSQL_PORT`：MySQL 端口（默认 `3306`）
- `MYSQL_ROOT_PASSWORD`：MySQL root 密码
- `MYSQL_USER` / `MYSQL_PASSWORD`：业务用户与密码（默认用户 `novel_reader`）
- `MYSQL_DATABASE`：数据库名（默认 `novel_reader`）
- `DATABASE_DSN`：后端连接 DSN（Compose 内默认连 `mysql:3306`）
- `UPLOAD_DIR`：容器内上传目录（默认 `/app/data/uploads`）
- `MAX_UPLOAD_BYTES`：上传大小限制（默认 `52428800`，即 50MB）
- `COVER_DIR`：封面存储目录（默认 `/app/data/uploads/covers`）
- `MAX_COVER_BYTES`：封面大小限制（默认 `10485760`，即 10MB）
- `DEFAULT_COVER_FILE`：默认封面文件路径（可选）。当书籍未上传封面时，`GET /api/books/{id}/cover` 会返回该文件内容。可写绝对路径，也可写相对路径（相对 `COVER_DIR`）。
- `JWT_SECRET`：JWT 密钥（本地也建议改掉）
- `ADMIN_USERNAME` / `ADMIN_PASSWORD`：管理员账号密码

说明：

- Compose 运行时，后端连接 MySQL 用的是容器网络地址 `mysql:3306`。
- 前端容器通过 nginx 反向代理 `/api` 到后端；开发时（`npm run dev`）通过 Vite proxy 代理 `/api`。

## 常用命令

- `make dev`：启动完整服务栈（Compose up + build）
- `make test`：运行后端与前端测试
- `make smoke`：运行最小 API/页面可用性检查
- `make docker-build`：构建后端与前端镜像

更多说明见：`docs/commands.md`。

## 常见问题

### 1) `docker-compose` / `docker compose` 启动报错（backend 一直重启，MySQL 1045）

现象：

- `docker compose ps` 看到 `backend` 处于 `Restarting`
- `docker compose logs backend` 出现：
  - `Error 1045 (28000): Access denied for user 'novel_reader' ...`

原因（最常见）：

- MySQL 数据卷（volume）已经用旧密码初始化过，后续你修改 `.env` 里的 `MYSQL_PASSWORD` 不会自动更新 MySQL 内部用户密码。

解决方式：

1. 保留数据：用 root 登录 MySQL 后，手动修改用户密码（适合你需要保留已有数据的情况）。
2. 不保留数据：删除 MySQL volume，让它按新的 `.env` 重新初始化。

删除 volume（会清空数据库数据）示例：

```bash
docker compose -f deploy/docker-compose.yml --env-file .env down --remove-orphans
docker volume rm novel-reader_mysql-data
docker compose -f deploy/docker-compose.yml --env-file .env up -d --build
```

### 2) 文档里写 `docker compose up`，但我在仓库根目录直接跑不工作

本项目 Compose 文件在 `deploy/docker-compose.yml`。推荐直接使用：

- `make dev`
- 或 `docker compose -f deploy/docker-compose.yml --env-file .env up --build`

## Harness 文档

项目采用 Harness Engineering 工作流，产品/架构/契约/命令都在 `docs/` 下：`docs/README.md`。
