# 后端 DDD 说明

本文档描述 Phase 10 后端在本仓库内采用的 DDD 分层方式、目录约定和职责边界。

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

## 当前分层

后端固定分为四层：

- `interface`
- `application`
- `domain`
- `infrastructure`

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
- `backoffice` 当前作为 `identity` 域下的后台运营账号与审核流程处理，不单独升级为顶层领域。

## 当前 application 组织方式

### identity

`command`：

- `register`
- `login`
- `update nickname`
- `change password`
- `admin login`
- `submit author application`
- `review author application`
- `promote reader to author`
- `create operator`
- `update operator`

`query`：

- `current user`
- `current operator`
- `current user application`
- `list author applications`
- `list front users`
- `list operators`

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

`command`：

- `upload book`
- `upload cover`

## 当前 domain service 设计方式

### Book

- `BookService`：创建 `Book`、`Chapter`、更新元数据载体，并负责“只有作者本人可以发布和修改自己书籍”的权限判断。

### Identity

- `IdentityService`：前台用户名规则、后台运营账号规则、密码规则、昵称规则、作者申请规则。

### Category

- `CategoryService`：分类对象构造与分类名规则。

## application 和 domain 的协作原则

- 前台认证和后台认证由 `identity` 应用层分开编排。
- 作者申请获批时，由 `application/identity` 同步协调“申请状态更新 + 用户角色升级”。
- 后台直接在用户列表提升作者时，由 `application/identity` 同步协调“用户角色升级 + 可选的申请状态回写”。
- `book` 域不直接依赖作者申请实体，只消费“当前 actor 是否具备作者能力”这一结果。
- 后台运营账号不参与书籍内容写流程。

## 特别约束

- Phase 10 默认允许清空旧数据并按新模型重建，不为旧 `user/admin` 语义保留兼容负担。
- 只有作者可以发布、上传、修改和删除自己的书籍与章节。
- 后台运营账号仅负责审核、用户治理、分类治理和站点设置。
