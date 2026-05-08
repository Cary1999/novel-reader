# 决策 0003：Harness-Driven Agent Team

## 状态

已接受，用于后续协作流程。

## 决策

项目采用 Harness-Driven Agent Team 工作流：

1. 需求或 Bug 进入。
2. 主 Agent 识别任务类型、风险和影响范围。
3. 主 Agent 调用合适的 skill 进行分析。
4. 主 Agent 生成或更新 Harness 草案。
5. 用户确认 Harness 草案。
6. 主 Agent 按确认后的 Harness 拆分角色任务。
7. 产品、UI、前端、后端、测试等角色 Agent 执行。
8. 主 Agent 集成、验证并回填 Harness。

## 理由

多 Agent 协作可以提高复杂任务的吞吐，但也会放大需求理解、接口契约和职责边界的偏差。先用 Harness 固化确认后的事实，再让角色 Agent 在明确边界内执行，可以提升可靠性和可追溯性。

## 约束

- 产品范围、API 契约、UI 状态和验收标准确认前，不进入并行编码。
- 角色 Agent 的输入必须来自确认后的 Harness。
- 主 Agent 对集成和最终验证负责。
- 重复问题必须回填到 Harness，必要时沉淀为 skill。

