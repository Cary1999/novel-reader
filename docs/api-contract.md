# API 契约

这是用于规划的初始契约。具体的 Kratos proto 文件将在用户确认后的后续阶段生成。

## 通用约定

- 所有请求和响应使用 JSON，上传接口除外。
- 登录后接口使用 `Authorization: Bearer <token>`。
- 分页参数从 `page=1` 开始。
- 时间字段使用 ISO 8601 字符串。
- 章节正文接口需要登录。
- 后台接口需要管理员角色。

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

## 认证

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
  "nickname": "reader1"
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
  "token": "session-or-jwt-token",
  "user": {
    "id": 1,
    "username": "reader1",
    "nickname": "reader1",
    "role": "user"
  }
}
```

### GET /api/auth/me

认证：

- 需要登录。

响应：

```json
{
  "id": 1,
  "username": "reader1",
  "nickname": "reader1",
  "role": "user"
}
```

### PATCH /api/auth/me

认证：

- 需要登录。

请求：

```json
{
  "nickname": "新的昵称"
}
```

响应：

```json
{
  "id": 1,
  "username": "reader1",
  "nickname": "新的昵称",
  "role": "user"
}
```

### PATCH /api/auth/password

认证：

- 需要登录。

请求：

```json
{
  "oldPassword": "password123",
  "newPassword": "new-password123"
}
```

响应：

```json
{
  "changed": true
}
```

## 书籍

### GET /api/categories

响应：

```json
{
  "items": [
    {
      "id": 1,
      "name": "玄幻"
    }
  ]
}
```

### GET /api/site-settings

认证：

- 公开可读。

响应：

```json
{
  "brandName": "阅卷书屋",
  "brandSubtitle": "Local Reading Archive",
  "brandIconUrl": "/api/site-settings/icon",
  "heroEyebrow": "发现好故事",
  "heroTitle": "一站式书屋",
  "heroDescription": "搜索、筛选、查看目录、登录后按章节阅读，也可以上传和维护你自己的 txt 小说。"
}
```

### GET /api/site-settings/icon

认证：

- 公开可读。

响应：

- 返回当前站点图标文件。
- 当管理员尚未上传图标时，返回系统默认图标（HTTP 200）。

### GET /api/books/search

查询参数：

- `q`：关键词，可选。
- `category`：分类，可选。
- `page`：页码，可选。
- `pageSize`：每页数量，可选。

响应：

```json
{
  "items": [
    {
      "id": 1,
      "title": "Example Novel",
      "author": "Example Author",
      "category": "Fantasy",
      "description": "Short excerpt",
      "chapterCount": 12,
      "latestChapterTitle": "Chapter 12",
      "recommendScore": 0,
      "coverUrl": "/api/books/1/cover",
      "createdAt": "2026-05-07T10:00:00+08:00"
    }
  ],
  "total": 1
}
```

### GET /api/books/recommendations

查询参数：

- `page`：页码，可选，默认 `1`。
- `pageSize`：每页数量，可选，默认 `20`。

说明：

- 按 `recommendScore` 从大到小排序。
- 同分时按 `createdAt desc`（或 `id desc`）保证顺序稳定。

响应：

```json
{
  "items": [
    {
      "id": 1,
      "title": "Example Novel",
      "author": "Example Author",
      "category": "Fantasy",
      "description": "Short excerpt",
      "chapterCount": 12,
      "latestChapterTitle": "Chapter 12",
      "recommendScore": 98,
      "coverUrl": "/api/books/1/cover",
      "createdAt": "2026-05-07T10:00:00+08:00"
    }
  ],
  "total": 1
}
```

### GET /api/books/{bookId}

响应：

```json
{
  "id": 1,
  "title": "Example Novel",
  "author": "Example Author",
  "category": "Fantasy",
  "description": "Full description",
  "chapterCount": 12,
  "recommendScore": 0,
  "coverUrl": "/api/books/1/cover"
}
```

### GET /api/books/{bookId}/chapters

响应：

```json
{
  "items": [
    {
      "id": 101,
      "index": 1,
      "title": "Chapter 1"
    }
  ]
}
```

### GET /api/books/{bookId}/chapters/{chapterId}

认证：

- 需要登录。

响应：

```json
{
  "id": 101,
  "bookId": 1,
  "index": 1,
  "title": "Chapter 1",
  "content": "Chapter content..."
}
```

### GET /api/books/{bookId}/cover

说明：

- 返回书籍封面图片的二进制内容。
- 如果该书没有上传封面，接口必须返回默认占位图（HTTP 200），避免前端出现 broken image。

响应：

- `Content-Type: image/jpeg` 或 `image/png` 或 `image/webp`
- Body: 图片 bytes

### POST /api/books/{bookId}/cover

认证：

- 需要登录。

权限：

- 管理员或书籍作者本人。

内容类型：

- `multipart/form-data`

字段：

- `file`：封面图片文件（JPG/PNG/WebP），最大 10MB。

响应：

```json
{
  "coverUrl": "/api/books/1/cover"
}
```

### PATCH /api/books/{bookId}

认证：

- 需要登录。

权限：

- 管理员或书籍作者本人。

请求：

```json
{
  "title": "书名",
  "categoryId": 1,
  "description": "简介",
  "recommendScore": 10
}
```

说明：

- `recommendScore` 仅允许管理员更新。
- 作者不可修改。

响应：

- 返回书籍详情对象（同 `GET /api/books/{bookId}`）。

## 登录用户作品

### GET /api/me/books

认证：

- 需要登录。

说明：

- 普通用户只返回自己的作品。

查询参数：

- `q`：关键词，可选。
- `categoryId`：分类 ID，可选。
- `page`：页码，可选。
- `pageSize`：每页数量，可选。

响应格式同 `GET /api/books/search`。

### POST /api/me/books

认证：

- 需要登录。

请求：

```json
{
  "title": "新小说",
  "categoryId": 1,
  "description": "简介"
}
```

响应：书籍对象，作者由当前登录用户决定。

### POST /api/me/books/upload

认证：

- 需要登录。

内容类型：

- `multipart/form-data`

字段：

- `title`：必填。
- `categoryId`：必填，必须是已有分类。
- `description`：可选。
- `file`：必填，格式为 `.txt`。

响应同管理员上传。

### PATCH /api/books/{bookId}

认证：

- 需要登录。

权限：

- 管理员或作者本人。

请求：

```json
{
  "title": "新书名",
  "categoryId": 1,
  "description": "新简介"
}
```

响应：书籍对象。

### DELETE /api/books/{bookId}

认证：

- 需要登录。

权限：

- 管理员或作者本人。

响应：

```json
{
  "deleted": true
}
```

### POST /api/books/{bookId}/chapters
### PATCH /api/books/{bookId}/chapters/{chapterId}
### DELETE /api/books/{bookId}/chapters/{chapterId}

认证：

- 需要登录。

权限：

- 管理员或作者本人。

章节新增、修改、删除规则与管理员章节管理一致。

## 管理后台

### POST /api/admin/books/upload

认证：

- 需要管理员角色。
- 最大文件大小为 50MB。

内容类型：

- `multipart/form-data`

字段：

- `title`：必填。
- `categoryId`：必填，必须是已有分类。
- `description`：可选。
- `file`：必填，格式为 `.txt`。

响应：

```json
{
  "bookId": 1,
  "chapterCount": 12,
  "uploadId": 1,
  "firstChapterTitle": "第一章 开始",
  "lastChapterTitle": "第十二章 结束"
}
```

### POST /api/admin/books

认证：

- 需要管理员角色。

请求同 `POST /api/me/books`。

说明：

- 管理员新建小说的作者展示为“系统”。

### GET /api/admin/books

认证：

- 需要管理员角色。

查询参数：

- `q`：关键词，可选。
- `category`：分类，可选。
- `page`：页码，可选。
- `pageSize`：每页数量，可选。

响应：

```json
{
  "items": [
    {
      "id": 1,
      "title": "Example Novel",
      "author": "Example Author",
      "category": "玄幻",
      "description": "简介",
      "chapterCount": 12,
      "latestChapterTitle": "第十二章",
      "createdAt": "2026-05-07T10:00:00+08:00"
    }
  ],
  "total": 1
}
```

### PATCH /api/admin/books/{bookId}

认证：

- 需要管理员角色。

请求：

```json
{
  "title": "新书名",
  "author": "新作者",
  "category": "新分类",
  "description": "新简介"
}
```

响应：

```json
{
  "id": 1,
  "title": "新书名",
  "author": "新作者",
  "category": "新分类",
  "description": "新简介",
  "chapterCount": 12,
  "latestChapterTitle": "第十二章"
}
```

### DELETE /api/admin/books/{bookId}

认证：

- 需要管理员角色。

响应：

```json
{
  "deleted": true
}
```

### POST /api/admin/books/{bookId}/chapters

认证：

- 需要管理员角色。

请求：

```json
{
  "title": "新章节标题",
  "content": "新章节正文"
}
```

响应：

```json
{
  "id": 101,
  "bookId": 1,
  "index": 13,
  "title": "新章节标题",
  "content": "新章节正文"
}
```

### PATCH /api/admin/books/{bookId}/chapters/{chapterId}

认证：

- 需要管理员角色。

请求：

```json
{
  "title": "修改后的章节标题",
  "content": "修改后的章节正文"
}
```

响应：

```json
{
  "id": 101,
  "bookId": 1,
  "index": 13,
  "title": "修改后的章节标题",
  "content": "修改后的章节正文"
}
```

### DELETE /api/admin/books/{bookId}/chapters/{chapterId}

认证：

- 需要管理员角色。

响应：

```json
{
  "deleted": true
}
```

## 管理员分类

### POST /api/admin/categories

认证：

- 需要管理员角色。

请求：

```json
{
  "name": "武侠"
}
```

### PATCH /api/admin/categories/{categoryId}

认证：

- 需要管理员角色。

请求：

```json
{
  "name": "新分类名"
}
```

### DELETE /api/admin/categories/{categoryId}

认证：

- 需要管理员角色。

响应：

```json
{
  "deleted": true
}
```

### PATCH /api/admin/site-settings

认证：

- 需要管理员角色。

请求：

```json
{
  "brandName": "阅卷书屋",
  "brandSubtitle": "Local Reading Archive",
  "heroEyebrow": "发现好故事",
  "heroTitle": "一站式书屋",
  "heroDescription": "搜索、筛选、查看目录、登录后按章节阅读，也可以上传和维护你自己的 txt 小说。"
}
```

响应：

- 返回完整站点设置对象（同 `GET /api/site-settings`）。

### POST /api/admin/site-settings/icon

认证：

- 需要管理员角色。

请求：

- `multipart/form-data`
- 字段：
  - `file`：必填，支持 `jpg` / `jpeg` / `png` / `webp` / `svg`

响应：

- 返回完整站点设置对象（同 `GET /api/site-settings`）。

若分类已被书籍使用，返回 `CONFLICT`。

## 错误格式

所有 API 应使用一致的错误响应：

```json
{
  "code": "UNAUTHORIZED",
  "message": "login required"
}
```

建议错误码：

- `BAD_REQUEST`
- `UNAUTHORIZED`
- `FORBIDDEN`
- `NOT_FOUND`
- `CONFLICT`
- `UPLOAD_INVALID_TYPE`
- `UPLOAD_TOO_LARGE`
- `PARSE_NO_CHAPTERS`
- `INTERNAL`
