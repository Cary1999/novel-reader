# 当前执行计划

当前指向：

- 已完成：`docs/exec-plans/phase-0-harness-skeleton.md`
- 已完成：`docs/exec-plans/phase-1-project-scaffold.md`
- 本轮基础验证已完成：`docs/exec-plans/phase-2-mvp-hardening.md`
- 已完成：`docs/exec-plans/phase-3-upload-limit-50mb.md`
- 已完成：`docs/exec-plans/phase-4-admin-book-management.md`
- 已完成：`docs/exec-plans/phase-5-user-authoring-account-categories-search-empty.md`
- 已完成：`docs/exec-plans/phase-6-ui-refresh-modern-reading-site.md`
- 已完成：`docs/exec-plans/phase-7-recommendations-and-covers.md`
- 已完成：`docs/exec-plans/phase-8-backend-ddd-refactor.md`
- 已完成：`docs/exec-plans/phase-9-site-brand-settings.md`
- 已完成：`docs/exec-plans/phase-10-three-portals-author-application.md`

## 当前状态

当前已完成项目骨架和 MVP 业务代码实现。

新需求“上传小说大小限制改为 50MB”已完成实现和验证。

新需求“后台小说和章节管理”已完成实现和验证。

Phase 2 的本轮 MVP 构建、接口和 smoke 验证已完成；浏览器视觉细节验收仍可继续补充。

新需求“用户创作、账号资料、分类治理和搜索空状态修复”已完成实现和验证。

新需求“现代中文阅读站风格 UI 改版”已完成实现与验证。

新需求“后端采用模块化 DDD 分层重构”已完成实现与验证。

新需求补充“新建小说时可直接设置封面”已按“两步提交、单次交互”的默认方案完成首轮实现，并通过前端与项目级测试。

新需求“管理员系统设置站点图标与品牌文案”已完成实现，并通过后端测试、前端测试、前端构建和项目级 `make test` 验证。

后续凡是涉及后端的需求，均需在 Harness 草案阶段主动补充 DDD 影响评估，再进入实现。

新需求“拆分读者端、作者端、后台端，并新增读者申请成为作者流程”已完成实现与当前轮验收；当前正式口径为“前台读者/作者体系”和“后台运营体系”完全分离，并引入 `super_admin` 创建审核人员、申请记录审核、用户列表直升作者、后台推荐值治理等能力；书籍发布与修改仅允许作者本人操作，后台不再承担内容正文编辑，且 Phase 10 允许清空旧数据并按新模型重建。

新需求“用户头像、首页合并搜索分页与我的书架”已完成实现并通过验收；当前确认口径为头像支持上传与展示、首页承载搜索结果且支持 `1-100` 自定义每页页数、`/search` 仅保留兼容跳转、书架支持分组/置顶/批量管理。

## 下一步

- 补充浏览器级移动端和桌面端视觉验收。
- 后续新需求继续先使用 `harness-agent-team` 更新 Harness 草案。
- 如需继续补充验收，可再补浏览器级移动端和桌面端视觉回归，但这不影响当前已完成阶段。
