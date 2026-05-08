# Phase 3：上传小说大小限制改为 50MB

## 状态

已完成。

## 任务类型

- 新功能
- UI 调整
- 后端配置变化
- 测试补强
- Harness 更新

## 背景

当前后台上传 `.txt` 小说的大小限制偏低。对于多数长篇小说来说，原容量不足，容易导致管理员上传失败。

## 目标

将后台 `.txt` 小说上传大小限制调整为 50MB，并确保后端限制、前端提示、配置示例、测试和 Harness 文档保持一致。

## 非目标

- 不新增分片上传。
- 不新增断点续传。
- 不改成本地文件系统之外的对象存储。
- 不调整上传文件类型，仍仅支持 `.txt`。
- 不改变管理员权限要求。
- 不改变章节解析规则。

## 影响范围

- 产品：后台上传能力的容量上限调整为 50MB。
- UI：管理员上传页的大小提示和前端校验调整为 50MB。
- 前端：`AdminUploadPage` 的 `MAX_FILE_SIZE` 和展示文案需要同步。
- 后端：默认 `MAX_UPLOAD_BYTES` 调整为 50MB。
- API：接口路径和请求响应不变；上传限制语义变化。
- 数据库：不需要 schema 变化，`uploads.file_size` 继续记录实际文件大小。
- 测试：补充或更新前后端上传大小限制相关测试。
- DevOps：`.env.example` 和 Compose 环境变量建议显式提供 `MAX_UPLOAD_BYTES=52428800`。
- 安全/权限：管理员权限不变；仍拒绝非 `.txt`、空文件和超过 50MB 的文件。

## 需要更新的 Harness 文件

- `docs/product-spec.md`
- `docs/ui-spec.md`
- `docs/data-model.md`
- `docs/security.md`
- `docs/api-contract.md`
- `docs/commands.md`
- `docs/exec-plans/current.md`
- `docs/harness/status.md`

## 契约变化

### API

不新增接口，不修改请求和响应结构。

语义变化：

- `POST /api/admin/books/upload` 的最大文件大小调整为 50MB。
- 超过 50MB 仍返回 `UPLOAD_TOO_LARGE`。

### 数据

无 schema 变化。

`uploads.file_size` 继续记录上传文件大小，50MB 范围内的值应可正常保存。

### UI 状态

管理员上传页需要同步：

- 文件选择前提示“最大 50MB”。
- 前端选择超过 50MB 的文件时显示“文件不能超过 50MB”。
- 其他状态不变：加载、错误、成功摘要、无权限。

## 角色分工

### 主 Agent

- 确认草案后组织实现。
- 集成后端、前端、DevOps 和测试修改。
- 运行验证命令。
- 回填 Harness 状态。

### 后端 Agent

- 修改默认上传限制为 50MB。
- 确保 `MAX_UPLOAD_BYTES` 环境变量仍可覆盖默认值。
- 更新或补充后端测试。

### 前端 Agent

- 修改上传页前端大小限制和提示文案。
- 更新或补充前端测试。

### DevOps Agent

- 在 `.env.example` 和 Compose 环境中显式配置 `MAX_UPLOAD_BYTES=52428800`。
- 确认 Docker 环境和本地环境行为一致。

### 测试 Agent

- 确认超过 50MB 的文件被拒绝。
- 确认 50MB 以内的 `.txt` 文件不被前端限制拦截。
- 确认 `make test`、`make docker-build`、`make smoke` 仍通过。

### Reviewer Agent

- 检查前后端限制是否一致。
- 检查 Harness 文档是否还有旧限制残留。

## 验收标准

- 后端默认上传限制为 50MB。
- `.env.example` 暴露 `MAX_UPLOAD_BYTES=52428800`。
- Docker Compose 将 `MAX_UPLOAD_BYTES` 传入后端容器。
- 前端管理员上传页提示最大 50MB。
- 前端选择超过 50MB 的文件时阻止提交并展示错误。
- `.txt` 上传权限和格式校验不变。
- 旧限制数字和旧字节表达式不再出现在 Harness、后端、前端、部署配置和环境示例中。
- `make test` 通过。
- `make docker-build` 通过。
- `make smoke` 通过。

## 验证命令

```bash
rg "旧限制关键词" docs backend frontend deploy .env.example
make test
make docker-build
make smoke
```

## 验证结果

- 旧限制残留检查：通过。
- `make test`：通过。
- `npm run build`：通过。
- Docker Compose 配置确认 `MAX_UPLOAD_BYTES=52428800`：通过。
- `make docker-build`：通过。
- `BASE_URL=http://localhost:8001 make smoke`：通过。

## 运行备注

验证时本机 `8000` 端口已有本地 Go 后端进程占用，因此临时使用 `8001` 启动新后端代码并完成 smoke 验证。要让 `8000` 上的服务应用 50MB 新限制，需要重启当前本地后端进程，或停止该进程后重新启动 Docker 后端容器。

## 待确认问题

无。默认按 50MB，即 `52428800` 字节执行。

## 确认门禁

用户已确认，已进入实现。
