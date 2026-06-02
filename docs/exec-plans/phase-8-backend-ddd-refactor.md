# Phase 8：后端模块化 DDD 分层重构

## 状态

已完成本轮后端内部结构重构与文档补齐。

## 任务类型

- 架构重构
- 后端分层调整
- Harness 更新
- 回归验证

## 背景

当前后端已经承载认证、搜索、阅读、作品管理、后台管理、分类治理和上传能力，但现有实现里存在“大 service”和“大 Store”的问题，业务边界和职责分层不够清晰。

本轮需要在不改变产品范围和既有对外 API 行为的前提下，把后端重构为更稳定、可维护的模块化 DDD 结构。

## 目标

1. 后端采用 `interface -> application -> domain -> infrastructure` 四层结构。
2. 顶层领域稳定为 `identity`、`book`、`category`、`upload`。
3. `chapter` 作为 `book` 域内核心实体，而不是独立顶层领域。
4. 移除业务对全局大仓储接口的直接依赖，改为按领域仓储契约协作。
5. 保持现有 API、数据库行为和 smoke 路径兼容。

## 非目标

- 不新增用户可见产品功能。
- 不改动前端交互。
- 不引入事件总线、MQ 或微服务拆分。
- 不主动修改 API 路径、响应字段和数据库表结构语义。

## 影响范围

- 后端目录结构。
- HTTP handler 装配方式。
- 应用层用例划分。
- MySQL 仓储实现组织方式。
- 后端单测和构建入口。
- Harness 架构文档和 ADR。

## 领域与用例划分

### 顶层领域

- `identity`
- `book`
- `category`
- `upload`

### 应用层用例

- `identity/auth`
- `book/search`
- `book/reading`
- `book/book-management`
- `book/chapter-management`
- `category/category-management`
- `upload/upload-import`
- `upload/cover-upload`

## 关键设计决策

- `reading` 保留为应用层用例，不作为领域命名。
- `authoring` 不作为独立领域。
- 跨领域调用默认由 application 层同步编排。
- 暂不引入 bus；若未来需要异步副作用，再引入事件机制。
- domain entity 与 MySQL model 必须分离，并通过 mapper 转换。
- application 负责 command/query、分页和用例输入规则，不把这些规则下放到 domain。
- MySQL SQL 仓储按领域放在 `infrastructure/data/mysql/repo/`，数据库模型放在 `infrastructure/data/mysql/model/`。

## 实施步骤

1. 已更新架构文档、ADR 和当前执行计划。
2. 已按 `entity/repository/service` 拆分 `identity`、`book`、`category`、`upload` 领域。
3. 已把 `identity`、`book`、`category` 的 application 重构为 `application.go + command/query handlers`。
4. 已把 MySQL model/repo 迁移到 `infrastructure/data/mysql/`。
5. 已把 MySQL migration/seed 与上传仓储职责拆开。
6. 已调整 HTTP、JWT、TXT parser 和应用服务 import。
7. 已补充后端 DDD 说明文档，并更新 README、架构文档与 ADR。
8. 已运行后端测试和项目测试。

## 验收标准

- 现有 HTTP API 路径和核心行为保持兼容。
- 后端不再通过单个大 `Store` 作为业务统一入口。
- 代码目录能直接体现领域边界和分层职责。
- `main.go` 不再承载业务语义。
- `make test` 通过。
- `make docker-build` 通过。
- 若环境允许，`make smoke` 继续通过。

## 验证命令

- `cd backend && env GOCACHE=/private/tmp/novel-reader-go-cache go test ./...`：通过。
- `make test`：通过。
- `cd frontend && npm run build`
- `make docker-build`
- `make smoke`
