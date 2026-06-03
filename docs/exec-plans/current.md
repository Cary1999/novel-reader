# 当前执行计划

当前指向：

- 已完成：`docs/exec-plans/phase-0-harness-skeleton.md`
- 已完成：`docs/exec-plans/phase-1-project-scaffold.md`
- 本轮基础验证已完成：`docs/exec-plans/phase-2-mvp-hardening.md`
- 已完成：`docs/exec-plans/phase-3-upload-limit-50mb.md`
- 已完成：`docs/exec-plans/phase-4-admin-book-management.md`
- 已完成：`docs/exec-plans/phase-5-user-authoring-account-categories-search-empty.md`
- 当前进行中：`docs/exec-plans/phase-6-ui-refresh-modern-reading-site.md`
- 当前进行中：`docs/exec-plans/phase-7-recommendations-and-covers.md`
- 当前进行中：`docs/exec-plans/phase-8-backend-ddd-refactor.md`

## 当前状态

当前已完成项目骨架和 MVP 业务代码实现。

新需求“上传小说大小限制改为 50MB”已完成实现和验证。

新需求“后台小说和章节管理”已完成实现和验证。

Phase 2 的本轮 MVP 构建、接口和 smoke 验证已完成；浏览器视觉细节验收仍可继续补充。

新需求“用户创作、账号资料、分类治理和搜索空状态修复”已完成实现和验证。

新需求“现代中文阅读站风格 UI 改版”已确认 Harness 草案，进入前端实现阶段。

新需求“后端采用模块化 DDD 分层重构”已确认领域方案，进入后端文档与代码重构阶段。

新需求补充“新建小说时可直接设置封面”已按“两步提交、单次交互”的默认方案完成首轮实现，并通过前端与项目级测试。

## 下一步

建议下一步进入：

- 重启当前占用 `8000` 端口的本地后端进程，让运行中的服务应用最新后台管理接口和 50MB 上传限制。
- 或释放 `8000` 端口后重启 Docker 后端容器。
- 补充浏览器级移动端和桌面端视觉验收。
- 后续新需求继续先使用 `harness-agent-team` 更新 Harness 草案。
- 可补充浏览器级移动端和桌面端视觉验收。
- 完成 Phase 6 前端改版后，补一次浏览器级桌面端和移动端视觉验收。
- 按 `docs/exec-plans/phase-8-backend-ddd-refactor.md` 推进后端分层重构，并在回归验证后更新 Harness 状态。

Phase 7 将在 Phase 6 基础上继续推进：

- 首页展示数量上限（搜索/最新入库最多 8 本）与“查看更多”跳转策略。
- 推荐榜单（推荐度字段 + 排序接口 + 前端展示）。
- 推荐度权限收紧为仅管理员可设置。
- 书籍封面上传与占位图返回策略。
- 新建小说表单支持预先选择封面，并在创建成功后立即完成封面上传或提示可重试的部分成功状态。
