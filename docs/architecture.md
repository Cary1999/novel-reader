# 架构说明

## 总览

系统是一个本地全栈小说阅读网站。

- 前端：Vite + React + TypeScript
- 后端：Go + Kratos
- 数据库：MySQL
- 文件存储：本地 `data/uploads`
- 构建与运行：Docker 和 Docker Compose

本项目后端采用“单体下的模块化 DDD 分层”：

- `interface`：承接 HTTP 输入输出、认证中间件和协议映射。
- `application`：承接用例编排、权限流程和跨领域协调。
- `domain`：承接稳定业务模型、仓储契约和核心规则。
- `infrastructure`：承接 MySQL、JWT、本地存储、TXT 解析和配置加载。

Phase 10 的核心变化不是继续扩展旧 `user/admin` 模型，而是引入双身份体系：

- 前台：`reader`、`author`
- 后台：`super_admin`、`reviewer`

同时明确：只有作者可以发布和修改书籍；后台运营账号不承担内容编辑职责。

## 高层组件

### Web 前端

职责：

- 读者端：公开的小说发现、搜索、详情、阅读、登录、注册和作者申请。
- 作者端：作者登录、我的作品、上传、新建和章节维护。
- 后台端：后台登录、作者申请审核、用户列表、分类管理、站点设置和审核人员管理。
- 只通过小型 API client 层调用后端 HTTP API。

### 后端 API

职责：

- 前台认证和后台认证。
- 昵称和密码修改。
- 作者申请和后台审核。
- 小说搜索和元数据查询。
- 章节查询。
- 作者上传校验。
- 作者自有作品管理。
- 后台用户治理、分类治理和站点设置。
- `.txt` 章节解析。
- MySQL 持久化。
- 本地文件存储协调。

### MySQL

职责：

- 前台用户和角色。
- 后台运营账号和角色。
- 作者申请记录。
- 书籍。
- 章节。
- 上传记录或源文件引用。
- 用户昵称和书籍作者归属。
- 用户头像路径。
- 用户书架分组和条目。

### 本地文件存储

职责：

- 在应用受控路径下保存原始上传的 `.txt` 文件。
- 保存封面图片文件。
- 保存用户头像文件。
- 绝不信任用户提供的路径。
- 生成的路径必须由服务端控制并规范化。
- 本地开发下若使用相对路径配置，路径应按仓库根目录解析，而不是按当前 shell 工作目录漂移。

## 后端领域划分

本项目当前采用以下顶层领域：

- `identity`：前台用户、后台运营账号、登录、密码、昵称、角色、作者申请。
- `book`：书籍、章节、作品归属、阅读元数据。
- `category`：分类治理和分类可用性约束。
- `upload`：文件上传、封面上传、上传记录、上传状态。
- `site`：站点品牌设置、首页品牌文案和站点图标引用。
- `bookshelf`：用户书架、分组、置顶和批量管理。

说明：

- `chapter` 是 `book` 域内核心实体，不单独升为顶层领域。
- `reading` 是应用层用例，不作为领域命名。
- `authoring` 是操作场景，不作为独立领域。
- `backoffice` 当前作为 `identity` 域中的后台运营账号与审核流程处理，不单独升级为顶层领域。

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

当前主要用例：

- `identity/front-auth`
- `identity/admin-auth`
- `identity/author-application`
- `identity/operator-management`
- `book/search`
- `book/reading`
- `book/book-management`
- `book/chapter-management`
- `category/category-management`
- `site/site-settings`
- `upload/upload-import`
- `upload/cover-upload`
- `identity/avatar-upload`
- `bookshelf/bookshelf-management`

### Domain 层

职责：

- 定义领域实体、值对象、领域服务和不变量。
- 定义仓储契约。
- 提供全局可复用的领域错误。

关键约束：

- `identity`：前台用户名唯一，后台运营账号用户名唯一，密码只保存哈希。
- `book`：只有作者本人可以发布和修改自己书籍；章节变更后要保持 `chapter_count` 和 `latest_chapter_title` 一致。
- `category`：分类名唯一，已被书籍使用的分类不可删除。
- `site`：站点设置始终按单例维护，图标引用由服务端控制。
- `upload`：上传路径由服务端生成，状态只能按允许的生命周期流转。
- `bookshelf`：同一用户同一本书只允许存在一条书架记录；默认分组必须存在，条目分组归属和置顶状态由 application 层协调。

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

## 领域协作规则

当前阶段采用“应用层同步编排”作为默认跨领域协作方式：

- 领域层不直接依赖其他领域的具体实现。
- 需要跨领域组合流程时，由 application 层统一协调。
- 单请求内需要强一致的流程优先使用同步调用和单事务边界。

示例：作者上传 TXT 并导入书籍

1. `upload` 保存源文件并记录上传记录。
2. `upload` 调用 TXT 解析器获取章节草稿。
3. `category` 校验分类存在。
4. `book` 创建书籍与章节。
5. `upload` 回写上传状态。

## 前端边界

前端应保持以下职责分离：

- 页面层：路由页面和页面级状态。
- 组件层：可复用 UI 组件和阅读器组件。
- API client 层：集中封装后端 HTTP 调用和认证 Header。
- Auth 状态层：分别保存前台登录态和后台登录态。
- 样式层：集中定义主题、字号和阅读页面排版变量。

前端不得绕过 API client 直接散落请求逻辑。

## 安全约束

- 密码必须哈希保存，不能明文存储。
- 作者上传 API 必须要求作者权限。
- 书籍和章节管理 API 必须要求作者权限，并校验作者归属。
- 后台审核与治理 API 必须要求后台运营账号权限。
- 分类写接口必须要求后台运营账号权限。
- 上传必须拒绝不支持的文件类型。
- 上传文件大小必须有限制。
- 文件路径必须由服务端生成。
- 阅读 API 不得暴露本地文件系统路径。
- 章节正文 API 必须要求登录。

## Docker 约束

- 后端必须支持 Docker 构建。
- Docker Compose 应能在本地运行后端、前端和 MySQL。
- 运行时配置应来自环境变量，或来自挂载到容器中的配置文件。
- Phase 10 默认允许清空旧数据并按新模型重建。
