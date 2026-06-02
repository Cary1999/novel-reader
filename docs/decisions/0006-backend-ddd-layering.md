# 0006：后端采用模块化 DDD 分层

## 状态

已接受。

## 背景

随着认证、搜索、阅读、用户作品管理、后台管理、分类治理和上传能力逐步增加，原有后端虽然已有 `service / repository / domain` 粗分层，但多个问题已经明显：

- 单个 `service` 同时承载用例编排、权限控制、参数校验、上传处理和领域规则。
- 单个 `Store` 仓储接口覆盖用户、书籍、章节、分类、上传，业务线混杂。
- `domain` 包更像共享结构集合，领域边界和业务语义不清晰。
- `main.go` 直接组装大颗粒实现，装配层和业务层耦合偏重。

项目需要在不改变现有对外 API 行为的前提下，把后端调整为更清晰、可维护、可测试的领域分层结构。

## 决策

- 后端采用“单体应用下的模块化 DDD 分层”。
- 分层固定为：`interface -> application -> domain -> infrastructure`。
- 顶层领域固定为：
  - `identity`
  - `book`
  - `category`
  - `upload`
- `chapter` 作为 `book` 域内核心实体，不单独升为顶层领域。
- `reading`、`book-management`、`chapter-management`、`upload-import` 作为应用层用例，不作为领域命名。
- 跨领域流程默认由 application 层同步编排，不默认引入消息总线。
- 只有当出现明确的异步副作用、重试或最终一致性需求时，才考虑进程内领域事件或外部 bus。
- 领域模型和数据库模型必须分离：
  - 领域实体放在 `domain/<domain>/entity/`。
  - 领域仓储契约放在 `domain/<domain>/repository/`。
  - 领域规则和策略放在 `domain/<domain>/service/`。
  - 应用层 command/query 放在 `application/<domain>/command/` 或 `application/<domain>/query/`。
  - MySQL 数据模型放在 `infrastructure/data/mysql/model/`。
  - MySQL 仓储实现放在 `infrastructure/data/mysql/repo/`。
- MySQL model 通过 `FromEntity` / `ToEntity` 与 domain entity 转换，不允许直接把数据库表模型当领域模型使用。
- application 不再保留“大 application service”作为默认模式，优先采用：
  - `application/<domain>/application.go`
  - `Commands`
  - `Queries`
  - 一用例一 handler 文件
- domain service 以领域为单位聚合，例如 `book.go`、`identity.go`、`category.go`。
- 当某个领域判断依赖已有领域状态时，允许 domain service 依赖 repository，先查询再判断。

## 影响

- 代码目录将从按技术层粗分调整为按“领域 + 分层”组合组织。
- 仓储接口按领域拆分，不再保留全局大 `Store` 作为业务依赖入口。
- HTTP handler 只面向应用层用例接口，不直接理解仓储细节。
- JWT、本地文件存储、TXT 解析和 MySQL 落库归位到基础设施层。
- `main.go` 只保留配置加载和依赖装配职责。
- `infrastructure/persistence/mysql` 只作为 schema、migration、seed 和 Store 装配入口，具体 SQL 仓储实现迁入 `infrastructure/data/mysql/repo`。

## 不做的事

- 本轮不主动修改对外 API 契约。
- 本轮不引入微服务。
- 本轮不为 DDD 引入额外中间件或消息系统。
