# 当前 Harness 状态

## 查询时间

2026-05-08

## 当前阶段

Phase 0 已完成：Harness 骨架和多角色协作规则已经建立。

Phase 1 已完成：项目骨架、MVP 业务代码、Docker/Makefile/Smoke 入口已经生成。

Phase 2 已完成本轮 MVP 接口、构建和 smoke 验证；浏览器视觉细节仍可继续做人工或浏览器工具验收。

Phase 3 已完成：后台上传大小限制已调整为 50MB。

Phase 4 已完成：后台小说和章节管理能力已经实现。

Phase 5 已完成并通过验证：用户创作、账号资料、分类治理和搜索空状态修复已落地。

当前 `make test`、前端 `npm run build`、`make docker-build` 和 `BASE_URL=http://localhost:8001 make smoke` 已通过。

## 已具备内容

- Agent 入口：`AGENTS.md`
- 产品规格：`docs/product-spec.md`
- UI 规格：`docs/ui-spec.md`
- 架构说明：`docs/architecture.md`
- API 契约：`docs/api-contract.md`
- 数据模型：`docs/data-model.md`
- 安全约束：`docs/security.md`
- 工程命令：`docs/commands.md`
- Harness 原则：`docs/harness/principles.md`
- 团队角色：`docs/team/roles.md`
- 团队流程：`docs/team/workflow.md`
- 交接模板：`docs/team/handoff-template.md`
- 当前计划入口：`docs/exec-plans/current.md`
- 决策记录：`docs/decisions/`

## 已收束默认决策

- 游客可以搜索和查看书籍详情，但不能阅读章节正文。
- 章节正文要求登录。
- 首页包含搜索、分类、新书和简单榜单。
- 阅读页支持基础字号、浅色/深色主题和章节目录。
- 后台上传支持 `.txt`，上传成功后展示解析摘要。
- 后台支持修改小说基础信息、管理章节和删除小说。
- 普通用户可以上传并管理自己的小说。
- 小说作者由系统用户引用计算，管理员作品显示为“系统”。
- 小说分类由管理员统一维护，小说表单只能选择已有分类。
- MVP 不做上传前解析预览。
- MySQL 保存用户、分类、书籍、章节和上传记录。
- 本地文件系统保存上传源文件。

## 已完成实现

- Go + Kratos 后端。
- MySQL schema、seed 和本地上传存储。
- 注册、登录、当前用户接口。
- 分类、搜索、书籍详情、章节列表、章节正文接口。
- 管理员 `.txt` 上传和章节解析。
- 管理员书籍列表、书籍基础信息编辑和整本删除。
- 管理员章节新增、章节编辑和章节删除。
- 用户昵称和修改密码接口。
- 普通用户作品列表、新建小说、上传小说和自有作品编辑删除。
- 管理员分类新增、重命名和删除。
- 搜索无结果空状态防崩溃处理。
- Vite + React + TypeScript 前端。
- 首页、登录、注册、搜索、详情、阅读、账号设置、我的作品、上传、后台小说管理和分类管理页面。
- Docker Compose、Dockerfile、Makefile 和 smoke 脚本。

## 验证结果

- `make test`：通过。
- 后端 `go build ./cmd/server`：通过。
- 前端 `npm run build`：通过。
- `docker compose -f deploy/docker-compose.yml --env-file .env config --quiet`：通过。
- `scripts/smoke.sh` 语法检查：通过。
- `make docker-build`：通过。
- `make smoke`：通过，覆盖健康检查、分类、搜索、搜索空结果、注册、登录、当前用户、昵称修改、密码修改、普通用户上传、普通用户作品管理、管理员 `.txt` 上传、后台书籍管理、章节新增/修改/删除、整本删除和管理员分类管理。

## 已修复集成问题

- 修复前端 Docker 构建时 `VITE_API_BASE_URL=/api` 导致请求变成 `/api/api/...` 的问题。
- Docker 构建现在使用空 `VITE_API_BASE_URL`，前端请求 `/api/...`，由 nginx 代理到 `backend:8000`。
- smoke 脚本会读取 `.env` 中的管理员账号密码，确保管理员上传链路被真实验证。
- 上传小说大小限制已调整为 50MB，后端默认、Docker Compose、前端校验、UI 文案和 Harness 文档已同步。
- 后台小说和章节管理已实现，所有管理接口要求管理员角色，普通用户会被拒绝。
- 后台章节变更使用事务刷新书籍章节数和最新章节标题；删除章节后重排同书后续章节序号。
- 搜索无结果时前端对 `items` 和 `total` 做空值兜底，稳定展示空状态。
- 账号和作品接口已引入 `nickname`、`owner_user_id`、作者归属校验和分类选择约束。

## 当前验证结果

- `make test`：通过。
- `npm run build`：通过。
- `make docker-build`：通过。
- `BASE_URL=http://localhost:8001 make smoke`：通过。
- 旧限制残留检查：通过。
- Phase 4 管理链路 smoke：通过，覆盖上传后修改书籍、新增章节、修改章节、删除章节、删除整本小说和删除后公开详情 404。
- Phase 5 smoke 脚本语法检查：通过。
- Phase 5 真实接口 smoke：通过。

## 残余风险

- 本轮没有使用浏览器自动截图工具验证所有页面在移动端和桌面端的视觉细节，后续可继续补充浏览器级 UI 验收。
- 浏览器级移动端和桌面端视觉验收尚未自动执行。

## 当前运行提醒

本机 `8000` 端口当前被一个本地 Go 后端进程占用。该进程如果是在改动前启动的，需要重启后才能应用 50MB 上传限制。Docker 后端容器也需要在释放 `8000` 端口后重新启动。

## 下一步建议

- 日常开发使用混合启动，Docker 只跑 MySQL，本地跑后端和前端。
- 交付前使用 `make dev` 和 `make smoke` 验证完整容器栈。
- 后续可补充浏览器级端到端测试和移动端视觉回归检查。

## 开发阶段启动建议

日常开发优先使用混合启动：Docker 只运行 MySQL，本地运行后端和前端。这样接口调试、日志观察、热更新和断点调试都更直接。全 Docker 启动用于交付前验收和环境一致性验证。
