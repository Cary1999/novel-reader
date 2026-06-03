# 架构说明

## 总览

系统是一个本地全栈小说阅读网站。

- 前端：Vite + React + TypeScript
- 后端：Go + Kratos
- 数据库：MySQL
- 文件存储：本地 `data/uploads`
- 构建与运行：Docker 和 Docker Compose

本项目后端从“按技术层粗分”调整为“单体下的模块化 DDD 分层”：

- `interface`：承接 HTTP 输入输出、认证中间件和协议映射。
- `application`：承接用例编排、权限流程和跨领域协调。
- `domain`：承接稳定业务模型、仓储契约和核心规则。
- `infrastructure`：承接 MySQL、JWT、本地存储、TXT 解析和配置加载。

本轮重构仅调整后端内部结构，不主动改变既有 HTTP API、数据库行为和 MVP 功能范围。

后续凡是涉及后端的新需求、Bug 修复或重构，都必须先做 DDD 影响评估，再进入实现：

- 它属于哪个顶层领域，还是只是现有领域内的一个用例。
- 它应落在 `interface`、`application`、`domain` 还是 `infrastructure`。
- 它是否引入新的仓储契约、事务边界、跨领域编排或基础设施适配器。
- 如果现有文档无法清晰回答以上问题，应先更新 Harness 或决策记录。

## 高层组件

### Web 前端

职责：

- 公开的小说发现和搜索。
- 登录和注册页面。
- 小说详情页和章节阅读页。
- 后台上传页面。
- 后台小说管理页面。
- 我的作品页面、账号设置页面和管理员分类管理页面。
- 只通过小型 API client 层调用后端 HTTP API。

### 后端 API

职责：

- 认证和角色检查。
- 昵称和密码修改。
- 小说搜索和元数据查询。
- 章节查询。
- 管理员上传校验。
- 管理员书籍和章节管理。
- 普通用户自有作品管理。
- 管理员分类治理。
- `.txt` 章节解析。
- MySQL 持久化。
- 本地文件存储协调。

### MySQL

职责：

- 用户和角色。
- 书籍。
- 章节。
- 上传记录或源文件引用。
- 管理员账号初始化状态。
- 用户昵称和书籍作者归属。

### 本地文件存储

职责：

- 在应用受控路径下保存原始上传的 `.txt` 文件。
- 保存封面图片文件。
- 绝不信任用户提供的路径。
- 生成的路径必须由服务端控制并规范化。
- 本地开发下若使用相对路径配置，路径应按仓库根目录解析，而不是按当前 shell 工作目录漂移。

## 后端领域划分

本项目当前采用以下顶层领域：

- `identity`：用户、身份、登录、密码、昵称、角色。
- `book`：书籍、章节、作品归属、阅读元数据。
- `category`：分类治理和分类可用性约束。
- `upload`：文件上传、封面上传、上传记录、上传状态。
- `site`：站点品牌设置、首页品牌文案和站点图标引用。

说明：

- `chapter` 是 `book` 域内核心实体，不单独升为顶层领域。
- `reading` 是应用层用例，不作为领域命名。
- `authoring` 是操作场景，不作为独立领域。

## 后端分层与职责

### Interface 层

职责：

- 接收 HTTP 请求并做协议级校验。
- 读取认证信息并映射为应用层可用的 actor/claims。
- 调用 application 用例。
- 将领域对象或应用结果映射为 JSON/文件响应。

约束：

- 不直接写 SQL。
- 不承载业务规则。
- 不直接访问本地文件系统。

### Application 层

职责：

- 编排业务用例。
- 协调跨领域流程。
- 处理权限流程、输入归一化和事务边界决策。
- 定义贴近用例的 command/query 输入输出模型。

当前目录约定：

- `application/<domain>/application.go` 负责聚合装配 `Commands` / `Queries`
- `application/<domain>/command/*.go` 一个文件一个 command handler
- `application/<domain>/query/*.go` 一个文件一个 query handler

当前主要用例：

- `identity/auth`
- `book/search`
- `book/reading`
- `book/book-management`
- `book/chapter-management`
- `category/category-management`
- `site/site-settings`
- `upload/upload-import`
- `upload/cover-upload`

约束：

- 不持有具体 MySQL 或文件系统实现。
- 通过领域仓储接口和基础设施适配器工作。
- 分页、页大小、HTTP 查询条件归一化等用例输入规则留在 application，不下放到 domain。

### Domain 层

职责：

- 定义领域实体、值对象、领域服务和不变量。
- 定义仓储契约。
- 提供全局可复用的领域错误。

关键约束：

- `identity`：用户名唯一，密码只保存哈希。
- `book`：章节必须从属于书；章节变更后要保持 `chapter_count` 和 `latest_chapter_title` 一致。
- `category`：分类名唯一，已被书籍使用的分类不可删除。
- `site`：站点设置始终按单例维护，图标引用由服务端控制。
- `upload`：上传路径由服务端生成，状态只能按允许的生命周期流转。

约束：

- domain 只保存领域内需要表达和判断的模型，不保存数据库表结构模型。
- 仓储契约按领域拆分，放在 `domain/<domain>/repository/`。
- 领域实体放在 `domain/<domain>/entity/`，领域规则和策略放在 `domain/<domain>/service/`。
- 当前 `domain/<domain>/service/` 以领域为单位聚合成单个 service 文件，例如：
  - `domain/book/service/book.go`
  - `domain/identity/service/identity.go`
  - `domain/category/service/category.go`

### Infrastructure 层

职责：

- MySQL 持久化实现。
- JWT token 签发和解析。
- 本地文件存储。
- TXT 章节解析。
- 配置加载与服务装配辅助。

约束：

- 不承载业务用例编排。
- 不决定产品权限规则。
- MySQL 数据模型放在 `infrastructure/data/mysql/model/`，不得与 domain entity 混用。
- MySQL 仓储实现放在 `infrastructure/data/mysql/repo/`，通过 `FromEntity` / `ToEntity` 完成数据库模型和领域实体转换。
- `infrastructure/persistence/mysql/` 仅保留 MySQL schema、migration、seed 和 Store 装配入口。

## 领域协作规则

当前阶段采用“应用层同步编排”作为默认跨领域协作方式：

- 领域层不直接依赖其他领域的具体实现。
- 需要跨领域组合流程时，由 application 层统一协调。
- 单请求内需要强一致的流程优先使用同步调用和单事务边界。

示例：上传 TXT 并导入书籍

1. `upload` 保存源文件并记录上传记录。
2. `upload` 调用 TXT 解析器获取章节草稿。
3. `category` 校验分类存在。
4. `book` 创建书籍与章节。
5. `upload` 回写上传状态。

当前不默认引入消息总线或跨进程事件总线。只有出现异步通知、失败重试、最终一致性投影等明确需求时，再考虑引入进程内领域事件或外部 bus。

## 当前仓库结构

后端目标结构如下：

```text
novel-reader/
  AGENTS.md
  docs/
  backend/
    cmd/
      server/
    internal/
      interface/
        http/
      application/
        identity/
          application.go
          command/
          query/
        book/
          application.go
          command/
          query/
        category/
          application.go
          command/
          query/
        site/
          application.go
          command/
          query/
        upload/
          query/
      domain/
        shared/
        identity/
          entity/
          repository/
          service/
        book/
          entity/
          repository/
          service/
        category/
          entity/
          repository/
          service/
        site/
          entity/
          repository/
          service/
        upload/
          entity/
          repository/
      infrastructure/
        data/
          mysql/
            model/
            repo/
        auth/
        config/
        parser/
        storage/
        persistence/
          mysql/
```

补充说明见 [backend-ddd.md](/Users/lhl/Documents/project/src/ai_test/novel-reader/docs/backend-ddd.md)。

## 前端边界

前端应保持以下职责分离：

- 页面层：路由页面和页面级状态。
- 组件层：可复用 UI 组件和阅读器组件。
- API client 层：集中封装后端 HTTP 调用和认证 Header。
- Auth 状态层：保存登录态、当前用户和权限判断。
- 样式层：集中定义主题、字号和阅读页面排版变量。

前端不得绕过 API client 直接散落请求逻辑。

## 安全约束

- 密码必须哈希保存，不能明文存储。
- 后台上传 API 必须要求管理员权限。
- 后台书籍和章节管理 API 必须要求管理员权限。
- 普通用户作品管理 API 必须要求登录，并校验作者归属。
- 分类写接口必须要求管理员权限。
- 上传必须拒绝不支持的文件类型。
- 上传文件大小必须有限制。
- 文件路径必须由服务端生成。
- 阅读 API 不得暴露本地文件系统路径。
- 章节正文 API 必须要求登录。

## Docker 约束

- 后端必须支持 Docker 构建。
- Docker Compose 应能在本地运行后端、前端和 MySQL。
- 运行时配置应来自环境变量，或来自挂载到容器中的配置文件。
