# Harness-Driven Agent Team 工作流

本流程用于处理新需求、Bug 修复和较大改动。小型文案或样式修改可以走轻量流程，但不得违反已确认的 Harness。

## 1. 需求或 Bug 进入

主 Agent 先判断：

- 这是新需求、Bug、重构、测试补强还是文档更新？
- 是否影响产品范围、UI、API、数据模型、部署或验证命令？
- 只要涉及后端，是否影响现有 DDD 分层、领域边界、仓储契约、应用用例编排或基础设施职责？
- 是否存在安全、权限、数据迁移或兼容性风险？
- 是否需要使用 skill 进行需求分析或 Bug 分析？

## 2. Skill 分析

当输入模糊、风险较高或影响跨模块时，主 Agent 应调用合适的 skill。

常见选择：

- 需求复杂：使用需求分析类 skill。
- Bug 复杂：使用 Bug 分析和修复规划类 skill。
- 多角色协作：使用 `harness-agent-team` skill。
- 重复流程：沉淀或调用项目专属 skill。

输出应是 Harness 草案，而不是直接代码。

## 3. Harness 草案

根据任务影响范围，主 Agent 更新或生成：

- `docs/product-spec.md`
- `docs/ui-spec.md`
- `docs/api-contract.md`
- `docs/architecture.md`
- `docs/commands.md`
- `docs/exec-plans/current.md`
- `docs/decisions/`

草案必须包含：

- 目标
- 非目标
- 影响范围
- 角色分工
- API 或数据契约变化
- UI 状态变化
- 若涉及后端：DDD 影响评估（领域归属、层级落点、仓储/事务/跨领域协作变化）
- 验收标准
- 必跑验证命令
- 待确认问题

## 4. 用户确认门禁

以下内容必须先获得用户确认，再进入实现：

- 产品范围和非目标。
- 涉及前后端交互的 API 契约。
- 涉及用户体验的页面和状态。
- 涉及数据持久化的数据模型变化。
- 任务拆分和验收标准。

如果只是小型 Bug 且不改变契约，主 Agent 可以记录轻量执行计划后直接修复。

## 5. 角色执行

用户确认后，主 Agent 按职责分派任务。

推荐顺序：

1. 产品和 UI 细化最终规格。
2. 若涉及后端，先确认 DDD 归位方案，再由后端、前端和测试在契约稳定后并行。
3. DevOps 在运行方式变化时参与。
4. Reviewer 在集成前或集成后审查。

角色 Agent 必须使用 `docs/team/handoff-template.md` 输出交接内容。

## 6. 主 Agent 集成

主 Agent 负责：

- 检查各角色输出是否符合 Harness。
- 处理接口、类型、路径和命令不一致。
- 运行统一验证命令。
- 修复集成问题。
- 更新执行计划状态。

## 7. 反馈回填

如果出现以下情况，必须回填 Harness：

- API 契约遗漏或不准确。
- UI 状态遗漏。
- 后端改动没有明确 DDD 落点，或实现开始偏离 `interface -> application -> domain -> infrastructure` 边界。
- 测试没有覆盖关键路径。
- 命令不可运行或含义不清。
- 同类 Bug 重复出现。
- 角色边界导致冲突。

回填方式包括：

- 更新文档。
- 补充测试。
- 补充脚本。
- 补充决策记录。
- 创建或改进 skill。
