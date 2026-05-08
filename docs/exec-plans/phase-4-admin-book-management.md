# Phase 4：后台小说和章节管理

## 状态

已完成。

## 任务类型

- 新功能
- UI 调整
- API/数据契约变化
- 后端业务能力扩展
- 测试补强
- Harness 更新

## 背景

当前后台只支持上传 `.txt` 小说。小说上传后，如果书名、作者、分类、简介、章节正文或章节目录存在错误，管理员无法修正；如果小说需要更新章节，也无法通过后台追加、修改或删除章节。

用户已确认本草案，当前阶段已进入实现并完成验证。

## 目标

为管理员提供小说维护能力：

1. 修改小说名称、作者、分类、简介等基础配置。
2. 修改指定章节的标题和正文。
3. 新增小说章节，用于更新小说。
4. 删除指定章节。
5. 删除整本小说。

## 非目标

- 不做普通用户投稿或作者工作台。
- 不做章节审核流。
- 不做草稿/发布双版本。
- 不做批量章节导入替换。
- 不做章节付费、评论、书架等非管理能力。
- 不做复杂富文本编辑器，MVP 使用纯文本编辑。

## 默认实现假设

- 所有管理接口都要求 `admin` 角色。
- 删除小说采用硬删除：从 `books` 删除，数据库级联删除 `chapters`。
- 删除整本小说后，前台搜索、详情和阅读都不可见。
- 上传源文件和 `uploads` 记录默认保留，用于追溯来源；不在本阶段删除本地源文件。
- 新增章节默认追加到末尾，`chapter_index = 当前最大序号 + 1`。
- 删除章节后，为保持目录连续，后续章节自动重新编号。
- 修改章节正文不改变 `chapter_index`。
- 修改或删除章节后，后端同步刷新 `books.chapter_count` 和 `books.latest_chapter_title`。
- 修改书籍分类时，如果分类不存在，则自动创建分类。

## 影响范围

- 产品：后台从“上传工具”扩展为“小说管理工具”。
- UI：新增后台小说管理列表、编辑书籍表单、章节管理列表、章节编辑/新增页面或弹窗、删除确认状态。
- 前端：新增管理员管理页面、API client 方法、表单校验、成功/错误状态。
- 后端：新增管理员书籍和章节 CRUD 接口、权限校验、事务处理、章节重排、统计字段刷新。
- API：新增后台管理接口，已有公开阅读接口不改变。
- 数据库：优先不改 schema；如需要软删除则必须新增字段和迁移。
- 测试：补充后端服务/handler 测试、前端 API/表单测试、smoke 扩展。
- DevOps：无新增服务；smoke 脚本可扩展管理链路验证。
- 安全/权限：所有修改/删除接口仅管理员可访问；普通用户和游客不可调用。

## 需要更新的 Harness 文件

- `docs/product-spec.md`
- `docs/ui-spec.md`
- `docs/api-contract.md`
- `docs/data-model.md`
- `docs/security.md`
- `docs/commands.md`
- `docs/exec-plans/current.md`
- `docs/harness/status.md`

## 契约变化

### API

新增后台书籍管理接口：

#### GET /api/admin/books

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

#### PATCH /api/admin/books/{bookId}

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

#### DELETE /api/admin/books/{bookId}

认证：

- 需要管理员角色。

响应：

```json
{
  "deleted": true
}
```

新增后台章节管理接口：

#### POST /api/admin/books/{bookId}/chapters

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

#### PATCH /api/admin/books/{bookId}/chapters/{chapterId}

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

#### DELETE /api/admin/books/{bookId}/chapters/{chapterId}

认证：

- 需要管理员角色。

响应：

```json
{
  "deleted": true
}
```

错误码新增或复用：

- `BAD_REQUEST`
- `UNAUTHORIZED`
- `FORBIDDEN`
- `NOT_FOUND`
- `CONFLICT`
- `INTERNAL`

### 数据

默认不新增数据库表和字段。

需要保证：

- 删除书籍时，`chapters` 通过外键级联删除。
- 新增章节时，`chapter_index` 自动追加。
- 删除章节后，后续章节重新编号，保持从 1 开始连续。
- 修改、新增、删除章节后，刷新 `books.chapter_count` 和 `books.latest_chapter_title`。

如果用户确认需要“可恢复删除”或“删除源文件”，则需调整数据模型：

- 可恢复删除：给 `books` 和 `chapters` 增加 `deleted_at`。
- 删除源文件：需要定义 `uploads` 与本地源文件的清理策略。

### UI 状态

新增或调整后台页面：

- 后台小说管理页：搜索、列表、分页、编辑入口、删除入口。
- 小说基础信息编辑：书名、作者、分类、简介。
- 章节管理页：章节列表、新增章节、编辑章节、删除章节。
- 删除确认：删除章节和删除整本小说必须二次确认。
- 成功状态：保存成功、章节新增成功、删除成功。
- 错误状态：校验错误、无权限、接口失败、章节不存在、书籍不存在。
- 加载状态：列表加载、保存中、删除中。

## 角色分工

### 主 Agent

- 确认草案后组织实现。
- 拆分后端、前端、测试和 Harness 更新任务。
- 集成并运行验证。
- 处理前后端契约偏差。

### 产品 Agent

- 确认删除策略、章节排序策略和源文件保留策略。
- 补充产品规格和验收标准。

### UI Agent

- 设计后台管理页面结构、编辑状态、删除确认和错误状态。
- 更新 `docs/ui-spec.md`。

### 后端 Agent

- 实现管理员书籍管理和章节管理接口。
- 实现事务、章节重排和统计字段刷新。
- 补充服务和 handler 测试。

### 前端 Agent

- 实现后台小说管理、书籍编辑、章节新增/编辑/删除页面。
- 扩展 API client。
- 补充前端测试。

### 测试 Agent

- 补充 API 权限、CRUD、章节重排、删除后不可访问的测试。
- 扩展 smoke，至少覆盖管理员新增章节和修改书籍基础信息。

### Reviewer Agent

- 检查权限边界、数据一致性、删除行为和 Harness 对齐。

## 验收标准

- 管理员可以修改小说书名、作者、分类和简介。
- 管理员可以修改指定章节标题和正文。
- 管理员可以为小说新增章节，新增章节追加到末尾。
- 管理员可以删除指定章节，删除后章节序号保持连续。
- 管理员可以删除整本小说，删除后前台搜索和详情不可访问。
- 普通用户和游客不能调用任何管理接口。
- 修改、新增、删除章节后，书籍章节数和最新章节标题正确更新。
- 删除整本小说后，对应章节正文不可继续访问。
- 前端后台页面覆盖加载、错误、空状态、保存成功、删除确认。
- `make test` 通过。
- `make docker-build` 通过。
- `make smoke` 通过或记录未覆盖项。

## 验证命令

```bash
make test
npm run build
make docker-build
make smoke
```

建议增加定向验证：

```bash
go test ./internal/service ./internal/httphandler
npm test -- --run
```

## 实现结果

- 后端已新增后台书籍列表、书籍基础信息修改、整本删除接口。
- 后端已新增章节追加、章节修改、章节删除接口。
- 后端使用事务处理章节追加、章节删除、章节重排和统计字段刷新。
- 前端已新增后台小说管理页，支持搜索、分页、基础信息编辑、章节新增/编辑/删除和整本删除。
- API client、导航、路由、样式和 smoke 脚本已同步更新。
- Harness 正式文档已同步产品、UI、架构、API、数据模型、安全和命令验证结果。

## 验证结果

- `make test`：通过。
- `go test ./...`：通过。
- `npm run build`：通过。
- `make docker-build`：通过。
- `BASE_URL=http://localhost:8001 make smoke`：通过。

Smoke 覆盖：

- 管理员 `.txt` 上传。
- 后台书籍列表。
- 修改书籍基础信息。
- 新增章节。
- 修改章节。
- 删除章节。
- 删除整本小说。
- 删除后公开详情返回 404。

## 已确认问题

1. 删除整本小说时，是否保留上传源文件和 `uploads` 记录？默认保留，用于追溯。
2. 删除章节后，是否自动重排章节序号？默认自动重排，保持目录连续。
3. 新增章节是否只追加到末尾？默认只追加到末尾，不支持插入到中间。
4. 删除是否需要可恢复？默认硬删除，不做回收站。

## 确认门禁

用户已确认，本阶段门禁已通过。
