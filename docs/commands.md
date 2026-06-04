# 工程命令

这些命令是本项目当前提供的确定性开发和验证入口。

## 已实现命令

```bash
make dev
```

本地运行前端、后端和 MySQL，用于开发。

```bash
make test
```

运行后端和前端测试。

```bash
make smoke
```

运行最小 API 和页面可用性检查。

```bash
make docker-build
```

构建后端和前端 Docker 镜像。

```bash
docker compose up
```

使用 MySQL 运行本地完整服务栈。

## Harness 要求

阶段完成前必须运行相关命令。若命令无法运行，需要在执行计划或 Harness 状态中记录阻塞原因。

Go 命令约束：

- 不允许把 `GOCACHE` 指到仓库工作区内，例如 `.cache/`。
- Go 编译和测试缓存统一使用系统默认缓存位置，不做项目内特殊化落盘。
- 如果当前执行环境因默认缓存或网络权限导致 Go 验证无法运行，应直接记录阻塞原因，而不是在项目目录中创建临时缓存目录绕过。

## 当前验证状态

- `make test`：已通过。
- `npm run build`：已通过。
- `make smoke`：已通过，覆盖健康检查、分类、搜索、搜索空结果、注册、登录、当前用户、昵称修改、密码修改、头像上传、作者角色升级、普通用户上传、普通用户作品管理、书架首页/分组详情/置顶/批量管理和管理员分类管理。
- `make docker-build`：已通过。
- 上传大小限制为 50MB 的版本已通过 `BASE_URL=http://localhost:8001 make smoke` 验证。
- 后台小说和章节管理版本已通过 `BASE_URL=http://localhost:8001 make smoke` 验证。
- 用户创作、账号资料、分类治理和搜索空状态版本已通过 `BASE_URL=http://localhost:8001 make smoke` 验证。

## 推荐开发方式

日常开发推荐使用混合启动：

- Docker 运行 MySQL。
- 本地运行后端，便于断点、日志和快速重启。
- 本地运行前端 Vite dev server，便于热更新。

全 Docker 启动更适合交付前验收。

推荐流程：

```bash
docker compose -f deploy/docker-compose.yml --env-file .env up -d mysql
cd backend && DATABASE_DSN="novel_reader:change-me-db-password@tcp(127.0.0.1:3306)/novel_reader?charset=utf8mb4&parseTime=True&loc=Local" HTTP_ADDR=:8000 go run ./cmd/server
cd frontend && npm run dev
```

补充说明：

- 本地从 `backend/` 目录执行 `go run ./cmd/server` 时，`.env` 里建议使用相对存储目录，例如 `data/uploads`、`data/uploads/covers` 和 `data/uploads/avatars`，这样会自动解析到仓库根目录下。
- Docker Compose 运行时可以继续把这些目录覆盖成容器内路径，例如 `/app/data/uploads`。
- 如需自定义位置，显式设置 `UPLOAD_DIR`、`COVER_DIR` 和 `AVATAR_DIR` 即可。
- 头像大小限制由 `MAX_AVATAR_BYTES` 控制，默认 5MB。
- Phase 10 目标端口规划：
  - 读者端：`3000`
  - 作者端：`3001`
  - 后台端：`3002`
  - 后端：`8000`
- Phase 10 默认允许清空旧数据并按新模型重建。
