# Phase 1：项目骨架生成

## 状态

已完成。

## 目标

生成小说阅读网站的可运行项目骨架。用户确认后，本阶段进一步扩展为实现 MVP 业务闭环。

## 范围

包含：

- Go + Kratos 后端骨架。
- Vite + React + TypeScript 前端骨架。
- MySQL 本地运行配置。
- Dockerfile 和 Docker Compose。
- Makefile 命令入口。
- 最小配置文件和示例环境变量。
- 最小 seed 数据设计。
- MVP 登录注册、搜索、阅读、后台上传业务代码。

不包含：

- 精细 UI 视觉打磨。
- 推荐算法。
- 付费、评论、书架等非 MVP 功能。

## 输入依据

- `AGENTS.md`
- `docs/product-spec.md`
- `docs/ui-spec.md`
- `docs/architecture.md`
- `docs/api-contract.md`
- `docs/data-model.md`
- `docs/security.md`
- `docs/commands.md`
- `docs/team/roles.md`
- `docs/team/workflow.md`

## 角色分工

### 主 Agent

- 创建 Phase 1 任务分解。
- 协调前端、后端、测试和 DevOps。
- 集成并运行验证。
- 更新 `docs/exec-plans/current.md` 状态。

### 后端 Agent

- 生成 Kratos 后端目录。
- 建立配置、路由、基础服务结构。
- 准备 proto/API 契约落地位置。
- 准备 MySQL 连接和迁移目录。

### 前端 Agent

- 生成 Vite + React + TypeScript 目录。
- 建立路由、API client、基础布局和页面占位。
- 准备登录态和权限路由结构。

### DevOps Agent

- 创建 Dockerfile、Docker Compose 和 Makefile。
- 配置 MySQL、本地上传目录和服务端口。

### 测试 Agent

- 建立最小测试目录和 smoke check 入口。
- 确认 `make test` 和 `make smoke` 的预期行为。

### Reviewer Agent

- 检查骨架是否符合 Harness。
- 检查是否提前实现了未确认业务。
- 检查命令是否清晰可执行。

## 计划生成的仓库结构

```text
novel-reader/
  AGENTS.md
  docs/
  backend/
    api/
    cmd/
    configs/
    internal/
    migrations/
    Dockerfile
  frontend/
    src/
    Dockerfile
  deploy/
    docker-compose.yml
  data/
    uploads/
  scripts/
  Makefile
```

## 验收标准

- 后端目录和前端目录存在。
- Docker Compose 能定义 MySQL、后端和前端服务。
- Makefile 包含 `dev`、`test`、`smoke`、`docker-build`。
- 后端、前端和部署配置可以通过统一命令发现。
- 不实现超出骨架范围的业务能力。
- 后续 Phase 2 可以在该骨架上实现认证。
- MVP 业务闭环可以通过后续 smoke 验证。

## 验证命令

Phase 1 完成后应运行或记录阻塞原因：

```bash
make test
make smoke
make docker-build
```

当前结果：

- `make test`：通过。
- 后端 `go build ./cmd/server`：通过。
- 前端 `npm run build`：通过。
- Docker Compose 配置校验：通过。
- `make docker-build`：阻塞，Docker daemon 未运行。

## 风险

- Kratos 生成工具或 Go 依赖可能需要网络访问。
- 前端依赖安装可能需要网络访问。
- Docker 构建可能受本机 Docker 状态影响。

## 确认门禁

用户已确认并已进入实现。本阶段完成后，后续改动继续按 Harness 草案和用户确认流程执行。
