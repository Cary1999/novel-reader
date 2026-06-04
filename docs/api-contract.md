# API 契约

这是 Phase 10 的规划契约。具体的 Kratos proto 文件将在实现过程中同步生成。

## 通用约定

- 所有请求和响应使用 JSON，上传接口除外。
- 前台登录后接口使用 `Authorization: Bearer <front-token>`。
- 后台登录后接口使用 `Authorization: Bearer <admin-token>`。
- 分页参数从 `page=1` 开始。
- 书库搜索和书架分页支持 `pageSize`，前台搜索结果允许 `1-100` 自定义页数。
- 时间字段使用 ISO 8601 字符串。
- 章节正文接口需要前台登录。
- 后台接口只接受后台运营账号 Token。

## 健康检查

### GET /health

响应：

```json
{
  "status": "ok"
}
```

兼容路径：

- `GET /healthz`
- `GET /api/health`

## 前台认证

### POST /api/auth/register

请求：

```json
{
  "username": "reader1",
  "password": "password123"
}
```

响应：

```json
{
  "userId": 1,
  "username": "reader1",
  "nickname": "reader1",
  "role": "reader"
}
```

### POST /api/auth/login

请求：

```json
{
  "username": "reader1",
  "password": "password123"
}
```

响应：

```json
{
  "token": "front-token",
  "user": {
    "id": 1,
    "username": "reader1",
    "nickname": "reader1",
    "role": "reader"
  }
}
```

### GET /api/auth/me

认证：

- 需要前台登录。

响应：

```json
{
  "id": 1,
  "username": "reader1",
  "nickname": "reader1",
  "role": "reader",
  "avatarUrl": "/api/users/1/avatar?v=1710000000000"
}
```

### POST /api/auth/me/avatar

认证：

- 需要前台登录。

请求：

- `multipart/form-data`
- 字段：`file`

响应：

- 与 `GET /api/auth/me` 一致，返回包含 `avatarUrl` 的当前用户信息。

### PATCH /api/auth/me

认证：

- 需要前台登录。

请求：

```json
{
  "nickname": "新的昵称"
}
```

### PATCH /api/auth/password

认证：

- 需要前台登录。

请求：

```json
{
  "oldPassword": "password123",
  "newPassword": "new-password123"
}
```

## 后台认证

### POST /api/admin/auth/login

请求：

```json
{
  "username": "reviewer1",
  "password": "password123"
}
```

响应：

```json
{
  "token": "admin-token",
  "operator": {
    "id": 1,
    "username": "reviewer1",
    "role": "reviewer",
    "status": "active"
  }
}
```

### GET /api/admin/auth/me

认证：

- 需要后台登录。

## 作者申请

### POST /api/author-applications

认证：

- 需要前台登录，且当前角色为 `reader`。

请求：

```json
{
  "penName": "青石",
  "reason": "希望开始连载自己的小说"
}
```

### GET /api/author-applications/me

认证：

- 需要前台登录。

## 公开查询

### GET /api/categories

### GET /api/site-settings

### GET /api/site-settings/icon

### GET /api/users/{userId}/avatar

### GET /api/books/search

### GET /api/books/recommendations

### GET /api/books/{bookId}

### GET /api/books/{bookId}/chapters

### GET /api/books/{bookId}/chapters/{chapterId}

认证：

- 需要前台登录。

### GET /api/books/{bookId}/cover

## 我的书架

### GET /api/me/bookshelf

认证：

- 需要前台登录。

请求参数：

- `groupId`：可选；传入分组 ID 时返回该分组内书籍，不传时返回书架首页中的未分组书籍。
- `page`：可选，默认 `1`。
- `pageSize`：可选，默认 `20`。

### GET /api/me/bookshelf/{bookId}

认证：

- 需要前台登录。

### POST /api/me/bookshelf/{bookId}

认证：

- 需要前台登录。

请求：

```json
{
  "groupId": 1
}
```

说明：

- 不传 `groupId` 或传 `0` 时，书籍加入书架首页而不是某个默认分组。

### PATCH /api/me/bookshelf/{bookId}

认证：

- 需要前台登录。

请求：

```json
{
  "groupId": 1,
  "pinned": true
}
```

说明：

- 传 `groupId: 0` 时可将书籍移回书架首页。

### DELETE /api/me/bookshelf/{bookId}

认证：

- 需要前台登录。

### GET /api/me/bookshelf/groups

认证：

- 需要前台登录。

说明：

- 仅返回用户创建的书架分组，不存在默认分组。

### GET /api/me/bookshelf/groups/{groupId}

认证：

- 需要前台登录。

### POST /api/me/bookshelf/groups

认证：

- 需要前台登录。

请求：

```json
{
  "name": "追更中"
}
```

### PATCH /api/me/bookshelf/groups/{groupId}

认证：

- 需要前台登录。

请求：

```json
{
  "name": "追更中",
  "sortOrder": 2
}
```

也支持：

```json
{
  "pinned": true
}
```

说明：

- `pinned` 可用于分组置顶和取消置顶。

### DELETE /api/me/bookshelf/groups/{groupId}

认证：

- 需要前台登录。

### POST /api/me/bookshelf/batch

认证：

- 需要前台登录。

请求：

```json
{
  "action": "move",
  "groupId": 2,
  "bookIds": [1, 2, 3]
}
```

## 作者端作品接口

### GET /api/me/books

认证：

- 需要作者登录。

### POST /api/me/books

认证：

- 需要作者登录。

### POST /api/me/books/upload

认证：

- 需要作者登录。

### PATCH /api/books/{bookId}

认证：

- 需要作者登录，且必须是该书作者本人。

### DELETE /api/books/{bookId}

认证：

- 需要作者登录，且必须是该书作者本人。

### POST /api/books/{bookId}/chapters

认证：

- 需要作者登录，且必须是该书作者本人。

### PATCH /api/books/{bookId}/chapters/{chapterId}

认证：

- 需要作者登录，且必须是该书作者本人。

### DELETE /api/books/{bookId}/chapters/{chapterId}

认证：

- 需要作者登录，且必须是该书作者本人。

### POST /api/books/{bookId}/cover

认证：

- 需要作者登录，且必须是该书作者本人。

## 后台审核与治理接口

### GET /api/admin/author-applications

认证：

- 需要后台登录。

### PATCH /api/admin/author-applications/{applicationId}

认证：

- 需要后台登录。

请求：

```json
{
  "decision": "approved",
  "reviewNote": "资料符合要求"
}
```

### GET /api/admin/users

认证：

- 需要后台登录。

### GET /api/admin/books

认证：

- 需要后台登录。

说明：

- 用于后台查看全站书籍并调整治理属性。
- 不用于后台编辑正文、章节或作者内容。

### PATCH /api/admin/books/{bookId}/recommend-score

认证：

- 需要后台登录。

请求：

```json
{
  "recommendScore": 18
}
```

### PATCH /api/admin/users/{userId}/author-role

认证：

- 需要后台登录。

请求：

```json
{
  "action": "promote_to_author",
  "reviewNote": "后台直接提升"
}
```

### POST /api/admin/operators

认证：

- 需要 `super_admin`。

### GET /api/admin/operators

认证：

- 需要 `super_admin`。

### PATCH /api/admin/operators/{operatorId}

认证：

- 需要 `super_admin`。

### POST /api/admin/categories

### PATCH /api/admin/categories/{categoryId}

### DELETE /api/admin/categories/{categoryId}

认证：

- 需要后台登录。

### PATCH /api/admin/site-settings

### POST /api/admin/site-settings/icon

认证：

- 需要后台登录。

## 说明

- Phase 10 默认允许清空旧数据并按新模型重建。
- 旧的后台书籍内容管理接口不再作为目标契约保留。
- 只有作者可以发布、上传、修改和删除自己的书籍与章节。
