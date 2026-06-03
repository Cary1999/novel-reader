# Phase 9：管理员系统设置（站点图标与品牌文案）

## 状态

已完成首轮实现与构建/测试验证。

## 任务类型

- 新需求
- UI 调整
- API/数据契约变化
- 数据库迁移
- 测试补强
- Harness 更新

## 背景

当前站点品牌信息写死在前端代码中：

- 顶部细条：`本地书库`、`发现、阅读、上传与维护你自己的小说`
- Header 品牌区：图标、`阅卷书屋`、`Local Reading Archive`
- 首页 Hero：`发现好故事`、`一站式书屋`、说明小字

这导致管理员无法通过后台调整站点对外展示的品牌名称、首页主标题和辅助文案。每次修改都需要改代码并重新发布，不符合“系统设置”预期。

## 目标

1. 管理员在后台新增“系统设置”能力。
2. 管理员可以配置站点品牌图标和首页/站点头部的核心展示文案。
3. 前台页面从后端读取站点设置，而不是继续使用硬编码默认文案。
4. 未配置或图标缺失时，前台仍能稳定展示默认值，不白屏、不出现 broken image。

## 非目标

- 不做完整 CMS 或页面搭建器。
- 不开放导航菜单名称、分类名称、按钮文案的任意配置。
- 不在本轮引入主题配色、字体、背景图等视觉皮肤系统。
- 不复制第三方品牌 Logo、视觉资产或专有文案。
- 默认不包含浏览器 favicon、`document.title`、SEO meta 的动态化，除非确认要一起纳入。

## 默认实现假设

- 本轮先覆盖截图中最核心、最稳定的品牌位：
  - 顶部细条左侧站点名
  - 顶部细条右侧辅助文案
  - Header 品牌图标
  - Header 品牌主名称
  - Header 品牌副标题
  - 首页 Hero 眉标题
  - 首页 Hero 主标题
  - 首页 Hero 描述小字
- 图标单独上传，文案单独保存，避免把文件上传和 JSON 保存耦合成一个复杂接口。
- 前台初始化时读取公开站点设置接口；若接口失败，回退到仓库内默认值。
- 管理后台当前不再展示“顶部站点名”和“顶部辅助文案”编辑项，后端保留字段用于兼容既有设置。
- 图标使用项目自有默认占位图标，不引用外部链接。

## 影响范围

- 产品：新增管理员“系统设置”能力，允许调整站点品牌展示。
- UI：管理后台增加“系统设置”标签页；前台顶部品牌区和首页 Hero 改为配置驱动。
- 前端：新增站点设置查询与保存逻辑；Layout、HomePage、AdminDashboardPage 增加设置入口与渲染。
- 后端：新增站点设置读写接口、图标上传接口与默认回退逻辑。
- API：新增公开读取和管理员更新站点设置接口。
- 数据库：新增站点设置单例表及图标路径字段。
- 测试：补充默认回退、管理员权限、图标上传校验、前台渲染回归测试。
- 安全/权限：仅管理员可修改；公开接口只返回前台渲染需要的安全字段。

## 需要更新的 Harness 文件

- `docs/product-spec.md`
- `docs/ui-spec.md`
- `docs/architecture.md`
- `docs/api-contract.md`
- `docs/data-model.md`
- `docs/security.md`
- `docs/exec-plans/current.md`

## 契约变化

### DDD 影响评估

- 领域归属：
  - 站点设置已确认落入独立 `site` domain；其核心是站点品牌配置，图标上传只是该领域的一种基础设施能力，不再混入 `upload` 领域。
- 分层落点：
  - `interface`：承接公开读取、管理员保存和图标上传 HTTP 接口。
  - `application`：编排读取站点设置、保存文案设置、上传图标并更新设置引用。
  - `domain`：定义站点设置对象、字段约束、默认值策略和图标引用规则。
  - `infrastructure`：实现 MySQL 持久化、图标文件存储和默认图标读取。
- 仓储与事务：
  - 需要新增站点设置仓储契约。
  - 图标上传与设置更新应由 application 层协调，避免把文件系统细节泄漏到 interface 或 domain。
- 跨领域协作：
  - 如需记录最近更新人，可读取 `identity` 的 actor 信息，但默认仍由 application 层同步编排，不引入事件机制。

### API

新增公开接口：

- `GET /api/site-settings`
  - 用途：前台读取当前站点品牌设置。
  - 权限：公开可读。
  - 响应建议包含：
    - `brandName`
    - `brandSubtitle`
    - `brandIconUrl`
    - `heroEyebrow`
    - `heroTitle`
    - `heroDescription`

新增管理员接口：

- `PATCH /api/admin/site-settings`
  - 用途：保存站点文案类设置。
  - 权限：仅管理员。
  - 请求建议包含：
    - `brandName`
    - `brandSubtitle`
    - `heroEyebrow`
    - `heroTitle`
    - `heroDescription`

- `POST /api/admin/site-settings/icon`
  - 用途：上传或替换站点品牌图标。
  - 权限：仅管理员。
  - 格式：`multipart/form-data`
  - 约束：
    - 仅允许 `jpg` / `jpeg` / `png` / `webp` / `svg`
    - 文件大小需要限制，并返回明确错误

图标返回策略建议：

- 前台通过 `brandIconUrl` 直接渲染。
- 若未上传图标，后端返回默认图标 URL 或稳定的占位图接口地址，不返回空字符串导致前台断裂。

### 数据

新增表建议：`site_settings`

字段建议：

- `id`：主键，固定单例行。
- `brand_name`
- `brand_subtitle`
- `brand_icon_path`
- `hero_eyebrow`
- `hero_title`
- `hero_description`
- `updated_by_user_id`
- `created_at`
- `updated_at`

约束建议：

- 全站只维护一份有效站点设置。
- 文案字段允许使用初始化默认值，但保存后应持久化。
- 图标路径必须由服务端生成，不信任用户文件名或路径。

### UI 状态

管理后台：

- 新增“系统设置”标签页。
- 初始加载态：显示正在读取当前站点设置。
- 保存成功态：显示明确成功反馈。
- 保存失败态：展示接口错误，不清空已编辑内容。
- 图标上传成功后立即预览。
- 图标类型/大小不合法时展示明确错误提示。

前台 Header / 首页：

- 正常态：展示后台配置的图标和文案。
- 默认回退态：配置尚未写入或接口失败时，回退到当前默认品牌文案。
- 无图标态：展示默认图标，不出现 broken image。

## 角色分工

- 主 Agent：更新 Harness 草案、等待确认、后续集成实现与验证。
- 产品 Agent：确认本轮可配置字段边界和非目标。
- UI Agent：明确后台“系统设置”页状态、前台品牌位回退态和图标预览交互。
- 前端 Agent：实现公开读取、后台编辑、图标上传和前台配置渲染。
- 后端 Agent：实现站点设置表、接口、权限和图标存储。
- 测试 Agent：补充接口、权限、回退和前台回归验证。
- Reviewer Agent：检查契约偏差、权限漏洞和默认回退是否完整。

## 验收标准

- 管理员可以在后台看到“系统设置”入口，并查看当前站点设置。
- 管理员可以修改品牌主名称、副标题、首页主标题和辅助小字，保存后刷新前台即可看到变化。
- 管理员可以上传并替换站点图标，保存后 Header 品牌区显示新图标。
- 游客和登录用户都能读取最新站点设置，但只有管理员可以修改。
- 未配置图标或配置读取失败时，前台仍使用默认图标和默认文案稳定渲染。
- 非管理员调用写接口时返回明确无权限错误。

## 验证命令

- `cd backend && go test ./...`
- `cd frontend && npm test -- --run`
- `cd frontend && npm run build`
- `make test`
- `make smoke`

补充说明：

- 已通过：
  - `cd backend && go test ./...`
  - `cd frontend && npm test -- --run`
  - `cd frontend && npm run build`
  - `make test`
- 尚未执行：
  - `make smoke`（需要运行中的本地服务栈）

## 待确认问题

1. 本轮是否只覆盖截图里的站点头部与首页 Hero 文案，不包含浏览器标签标题和 favicon？
2. 顶部细条右侧那句“发现、阅读、上传与维护你自己的小说”是否也纳入可配置范围？
3. 图标格式是否接受 `svg`？推荐接受，便于站点品牌图标保持清晰。
4. 是否需要在页脚同步展示可配置站点名称？默认建议先不动页脚免责声明。

## 实现结果摘要

- 后端新增独立 `site` domain，包含站点设置 entity、repository、service、application command/query 与 MySQL 单例表。
- 新增公开接口 `GET /api/site-settings`、`GET /api/site-settings/icon`，以及管理员接口 `PATCH /api/admin/site-settings`、`POST /api/admin/site-settings/icon`。
- 前端新增全局站点设置 provider，驱动顶部品牌区与首页 Hero。
- 管理后台新增“系统设置”标签页，当前支持品牌名称、副标题、首页主视觉文案和站点图标上传。
