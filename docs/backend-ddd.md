# 后端 DDD 说明

本文档描述当前后端在本仓库内采用的 DDD 分层方式、目录约定和职责边界。

目标不是追求“形式上的 DDD 名词齐全”，而是让代码满足以下要求：

- 业务入口清晰。
- 领域边界清晰。
- 领域模型与数据库模型分离。
- 同类代码放在一起，便于维护和扩展。

同时，本文件不只是描述“当前结构”，也是后续所有后端需求进入实现前的检查基线。

## 后端需求进入前的必答清单

凡是需求、Bug、重构或运维调整涉及后端，都必须先回答：

1. 这是哪个 domain 的能力变化。
2. 这是新用例、现有用例扩展，还是基础设施适配变化。
3. 代码应该落在哪一层：
   - `interface`
   - `application`
   - `domain`
   - `infrastructure`
4. 是否需要新增或调整：
   - 领域实体
   - 仓储契约
   - domain service / policy
   - application command/query handler
   - MySQL model / repo / migration
5. 是否存在跨领域协作；如果有，是否仍应由 application 层同步编排。

如果这些问题没有答案，就不应直接开始写后端实现。

## 当前分层

后端固定分为四层：

- `interface`
- `application`
- `domain`
- `infrastructure`

### interface

位置：

- `backend/internal/interface/http`

职责：

- 承接 HTTP 请求与响应。
- 做协议层解析，例如 `json`、`multipart/form-data`、路径参数读取。
- 从认证中间件上下文提取 actor。
- 调用 application handler。

约束：

- 不写领域规则。
- 不直接写 SQL。
- 不直接操作数据库模型。

### application

位置：

- `backend/internal/application/<domain>`

职责：

- 作为业务入口。
- 编排一个完整用例。
- 处理用例级输入校验和归一化。
- 调用 domain service 和 repository 完成业务流程。
- 负责把仓储错误转换成用例语义错误。

结构约定：

- `application/<domain>/application.go`
- `application/<domain>/command/*.go`
- `application/<domain>/query/*.go`

说明：

- `command` 表示会改变状态的用例。
- `query` 表示只读用例。
- 一个 handler 对应一个明确用例，不再保留“大 application service”。

### domain

位置：

- `backend/internal/domain/<domain>`

职责：

- 定义领域实体。
- 定义领域仓储契约。
- 定义领域规则、策略、工厂和领域行为。

结构约定：

- `domain/<domain>/entity/*.go`
- `domain/<domain>/repository/*.go`
- `domain/<domain>/service/*.go`

当前 service 目录按领域聚合：

- `book/service/book.go`
- `identity/service/identity.go`
- `category/service/category.go`

示例：

- `book/service/book.go`
- `identity/service/identity.go`
- `category/service/category.go`

边界原则：

- 如果规则只是 `page/pageSize`、字符串裁剪、协议字段兜底，这属于 application。
- 如果规则是“某个领域对象能不能做某件事”，这属于 domain。
- 如果做这个判断必须先拿到已有领域状态，那么 domain service 可以依赖 repository 先查询再判断。

### infrastructure

位置：

- `backend/internal/infrastructure/...`

职责：

- MySQL 持久化实现。
- 本地文件存储。
- JWT 签发与校验。
- TXT 解析。
- 配置加载。

MySQL 相关约定：

- 数据库模型在 `backend/internal/infrastructure/data/mysql/model`
- 仓储实现在 `backend/internal/infrastructure/data/mysql/repo`
- migration/seed/store 装配入口在 `backend/internal/infrastructure/persistence/mysql`

## 领域划分

当前顶层领域固定为：

- `identity`
- `book`
- `category`
- `site`
- `upload`

说明：

- `chapter` 是 `book` 域内实体，不单独升级为顶层领域。
- `reading` 是用例，不作为领域命名。
- `authoring` 是操作场景，不作为独立领域。

## 目录示例

```text
backend/internal/
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
      service.go
      query/
  domain/
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
    persistence/
      mysql/
```

## 领域模型与数据库模型分离

这是当前重构里的硬约束。

- `domain/entity` 里的对象是领域实体。
- `infrastructure/data/mysql/model` 里的对象是数据库模型。

两者不能混用。

转换方式：

- 数据库模型提供 `FromEntity`
- 数据库模型提供 `ToEntity`

目的：

- 避免业务代码依赖表结构细节。
- 避免领域模型被数据库字段污染。
- 后续更换持久化实现时，应用层和领域层不需要跟着改。

## 当前 application 组织方式

### identity

`command`：

- `register`
- `login`
- `update nickname`
- `change password`

`query`：

- `current user`

### book

`command`：

- `create book`
- `update book`
- `delete book`
- `add chapter`
- `update chapter`
- `delete chapter`

`query`：

- `search books`
- `list recommended books`
- `get book`
- `list chapters`
- `get chapter`
- `list owned books`

### category

`command`：

- `create category`
- `update category`
- `delete category`

`query`：

- `list categories`

### site

`command`：

- `update site settings`
- `update site icon`

`query`：

- `get site settings`

### upload

当前仍是较小的应用服务：

- `upload book`
- `upload cover`

它已经使用 `book` 域的 `factory/policy`，后续如果继续扩展上传流程，也可以拆成 command/query handler。

## 当前 domain service 设计方式

### Book

- `BookService`：创建 `Book`、`Chapter`、更新元数据载体，并负责书籍管理权限判断。

### Identity

- `IdentityService`：注册用户构造、用户名规则、密码规则、昵称规则。

### Category

- `CategoryService`：分类对象构造与分类名规则。

## application 和 domain 的协作原则

典型流程：

1. `application` 接收用例输入。
2. `application` 做输入归一化。
3. `application` 调用 `domain` 的 factory/policy。
4. `application` 通过 repository 持久化或查询。
5. `application` 返回结果给 `interface`。

例如创建书籍：

1. `application/book/command/create_book_handler.go` 接收输入。
2. 校验分类是否存在。
3. 调用 `BookService.NewBook(...)` 创建领域对象。
4. 调用 `BookRepository.CreateBook(...)` 持久化。

例如书籍管理权限判断：

1. `application` 调用 `BookService.CanManageBook(ctx, bookID, actor)`。
2. `BookService` 先查出书籍。
3. `BookService` 再根据 owner/admin 规则做判断。

## 跨领域调用

当前默认采用 application 层同步编排，不默认引入消息总线。

原因：

- 当前需求以单请求内强一致为主。
- 引入 bus 会增加复杂度和调试成本。

只有在出现以下需求时再考虑事件机制：

- 异步副作用。
- 重试。
- 最终一致性投影。
- 明确的解耦收益大于复杂度。

## 当前仍保留的现实折中

- `upload` 还没有完全拆成 command/query。
- `Store` 仍作为装配入口存在于 `persistence/mysql`，但业务代码不再把它当“大业务仓储”使用，而是通过领域 repository 接口依赖它。
- 某些仓储还未全部迁到统一的 GORM 风格，当前以“边界清晰优先，技术实现一致性其次”为原则。

## 后续演进建议

- 当某个领域出现更多“依赖已有状态的复杂规则”时，继续从 application 收回到 domain policy 或 domain behavior。
- 当某个 application 目录再次出现“大 handler / 大 service”时，继续按单用例拆分。
- 当跨领域异步需求出现时，再评估是否引入进程内领域事件或外部 bus。
- 当新需求无法自然落入现有 DDD 约束时，先补 Harness 或决策记录，再动代码，而不是先实现后解释。
