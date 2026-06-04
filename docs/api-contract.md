# API 契约

这是 Phase 10 的规划契约。具体的 Kratos proto 文件将在实现过程中同步生成。

## 通用约定

- 所有请求和响应使用 JSON，上传接口除外。
- 前台登录后接口使用 `Authorization: Bearer <front-token>`。
- 后台登录后接口使用 `Authorization: Bearer <admin-token>`。
- 分页参数从 `page=1` 开始。
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
  "role": "reader"
}
```

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

### GET /api/books/search

### GET /api/books/recommendations

### GET /api/books/{bookId}

### GET /api/books/{bookId}/chapters

### GET /api/books/{bookId}/chapters/{chapterId}

认证：

- 需要前台登录。

### GET /api/books/{bookId}/cover

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
