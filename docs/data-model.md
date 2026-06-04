# 数据模型

本文件定义 Phase 10 的 MySQL 数据模型和本地文件存储规则。

## 表清单

### users

用途：保存前台用户账号和角色。

字段：

- `id`：主键。
- `username`：唯一用户名。
- `nickname`：昵称，注册时默认等于用户名。
- `password_hash`：密码哈希。
- `role`：`reader` 或 `author`。
- `created_at`：创建时间。
- `updated_at`：更新时间。

约束：

- `username` 唯一。
- `password_hash` 不允许为空。
- `role` 默认 `reader`。

### operators

用途：保存后台运营账号。

字段：

- `id`：主键。
- `username`：唯一用户名。
- `password_hash`：密码哈希。
- `role`：`super_admin` 或 `reviewer`。
- `status`：`active` 或 `disabled`。
- `created_by_operator_id`：创建人 ID，可为空。
- `created_at`：创建时间。
- `updated_at`：更新时间。

约束：

- `username` 唯一。
- `password_hash` 不允许为空。
- 系统初始化时至少需要 1 个 `super_admin`。

### categories

用途：保存小说分类。

字段：

- `id`：主键。
- `name`：分类名称。
- `created_at`：创建时间。

约束：

- `name` 唯一。

### site_settings

用途：保存全站品牌文案和站点图标引用。

字段：

- `id`：主键，固定单例行。
- `brand_name`：品牌主名称。
- `brand_subtitle`：品牌副标题。
- `brand_icon_path`：站点图标相对路径，可为空。
- `hero_eyebrow`：首页 Hero 眉标题。
- `hero_title`：首页 Hero 主标题。
- `hero_description`：首页 Hero 描述小字。
- `updated_by_operator_id`：最近更新后台运营账号 ID，可为空。
- `created_at`：创建时间。
- `updated_at`：更新时间。

约束：

- 全站只维护一份有效站点设置。
- 文案字段由服务端校验长度和必填。
- `brand_icon_path` 由服务端生成，不接受客户端直接写入绝对路径。

### books

用途：保存小说元数据。

字段：

- `id`：主键。
- `title`：书名。
- `author`：作者展示名。
- `owner_user_id`：作者用户 ID，引用 `users.id`。
- `category_id`：分类 ID，可为空。
- `description`：简介。
- `chapter_count`：章节数。
- `latest_chapter_title`：最新章节标题。
- `recommend_score`：推荐度，整数，默认 `0`。
- `cover_path`：封面文件相对路径，可为空。
- `source_upload_id`：来源上传记录 ID，可为空。
- `created_at`：创建时间。
- `updated_at`：更新时间。

索引：

- `title`
- `owner_user_id`
- `category_id`
- `recommend_score`
- `created_at`

### chapters

用途：保存章节元数据和正文。

字段：

- `id`：主键。
- `book_id`：书籍 ID。
- `chapter_index`：章节序号，从 1 开始。
- `title`：章节标题。
- `content`：章节正文。
- `created_at`：创建时间。

约束：

- 同一本书内 `chapter_index` 唯一。
- 删除书籍时应删除对应章节。
- 删除指定章节后，同一本书中后续章节序号应重新编号，保持连续。
- 新增章节默认追加到当前最大 `chapter_index` 之后。

索引：

- `(book_id, chapter_index)`

### uploads

用途：保存作者上传记录和源文件引用。

字段：

- `id`：主键。
- `user_id`：上传作者 ID。
- `original_filename`：原始文件名，仅用于展示。
- `stored_path`：服务端生成的相对存储路径。
- `file_size`：文件大小。
- `status`：`uploaded`、`parsed` 或 `failed`。
- `error_message`：解析失败原因，可为空。
- `created_at`：创建时间。

约束：

- `stored_path` 由服务端生成，不来自用户输入。
- API 不向普通用户暴露 `stored_path`。

### author_applications

用途：保存作者申请记录。

字段：

- `id`：主键。
- `user_id`：前台用户 ID。
- `status`：`pending`、`approved` 或 `rejected`。
- `pen_name`：可为空。
- `reason`：可为空。
- `review_note`：可为空。
- `reviewed_by_operator_id`：审核运营账号 ID，可为空。
- `reviewed_at`：审核时间，可为空。
- `created_at`：创建时间。
- `updated_at`：更新时间。

约束：

- 同一用户同一时间最多一条 `pending` 申请。

## 初始数据

项目首次启动或 seed 时应包含：

- 一个后台 `super_admin` 账号，用户名和初始密码来自环境变量或本地开发配置。
- 若干基础分类，例如玄幻、都市、科幻、历史、游戏。
- 示例前台读者、作者账号和示例书籍可按 smoke 需要初始化。

## 迁移规则

- Phase 10 默认允许清空旧数据并按新模型重建。
- 不要求兼容旧的 `user/admin` 角色语义。
- 若需要保留示例数据，应按新角色模型重新初始化，而不是做旧数据迁移映射。

## 本地文件存储

上传源文件保存到：

```text
data/uploads/
```

规则：

- 目录由服务端创建。
- 默认相对路径按仓库根目录解析，保证本地 `go run` 与 Docker 启动写入同一逻辑位置。
- 文件名由服务端生成，避免使用原始文件名作为路径。
- 仅允许作者上传 `.txt` 文件。
- 上传大小限制在配置中定义，MVP 默认限制为 50MB。
- 作者删除整本小说时删除 `books` 和级联 `chapters`；上传源文件和 `uploads` 记录是否保留由实现阶段明确。

封面文件保存到：

```text
data/uploads/covers/
```

规则：

- 默认相对路径按仓库根目录解析，保证本地 `go run` 与 Docker 启动写入同一逻辑位置。
- 文件名由服务端生成，避免使用原始文件名作为路径。
- 允许 `jpg`/`jpeg`、`png`、`webp`。
- 最大尺寸 10MB。
- 如果书籍没有封面，封面读取接口必须返回服务端占位图（HTTP 200），而不是 404。
