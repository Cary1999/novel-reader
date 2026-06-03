# Agent 入口指南

本仓库采用 Harness Engineering 工作流进行开发。

目标不是依赖一段很长的提示词，而是让仓库本身成为产品意图、架构、契约、约束、执行计划和验证命令的事实来源。

## 工作模型

人类掌舵，Agent 执行。

人类负责决定产品范围、技术取舍和验收标准。Agent 负责读取仓库内的上下文，进行小范围、可验证的修改，运行验证命令，并在重复失败暴露约束缺口时更新 Harness。

## 上下文地图

开始修改前，优先阅读以下文件：

- 产品范围：`docs/product-spec.md`
- UI 规格：`docs/ui-spec.md`
- 技术架构：`docs/architecture.md`
- API 契约：`docs/api-contract.md`
- 数据模型：`docs/data-model.md`
- 安全约束：`docs/security.md`
- 工程命令：`docs/commands.md`
- Harness 原则：`docs/harness/principles.md`
- Harness 状态：`docs/harness/status.md`
- 团队角色：`docs/team/roles.md`
- 团队流程：`docs/team/workflow.md`
- 交接模板：`docs/team/handoff-template.md`
- 当前任务入口：`docs/exec-plans/current.md`
- 决策记录：`docs/decisions/`

## 项目意图

构建一个本地小说阅读网站。产品结构和阅读体验可以参考起点读书，但不得复制其品牌、Logo、视觉资产或专有内容。

MVP 能力：

- 注册和登录。
- 搜索小说。
- 按章节阅读小说。
- 在后台上传 `.txt` 小说。
- 使用 MySQL 存储结构化数据。
- 使用本地文件系统保存上传源文件。
- 支持 Docker 构建和本地运行。

## 必选技术栈

- 后端：Go + Kratos
- 前端：Vite + React + TypeScript
- 数据库：MySQL
- 上传文件：本地文件系统
- 构建与运行：Docker 和 Docker Compose

## Agent 规则

- 在用户确认 Harness 骨架之前，不生成应用代码。
- 需求或 Bug 进入后，主 Agent 应先判断是否需要 skill 分析，再生成或更新 Harness 草案。
- 只要需求涉及后端，主 Agent 必须主动评估该需求的 DDD 影响：领域归属、分层落点、是否新增或调整 repository 契约、是否需要补充决策记录或更新 `docs/backend-ddd.md` / `docs/architecture.md`。
- 未经用户确认的 Harness 草案不得作为代码实现依据。
- 多角色 Agent 可以并行执行，但产品范围、接口契约、UI 状态和验收标准必须先串行确认。
- 优先采用契约先行。行为变化时，先更新 API 和数据契约，再进行实现。
- 涉及后端实现时，默认要求实现符合既有模块化 DDD 分层，不得绕过 `interface -> application -> domain -> infrastructure` 的边界约束。
- 计划和文档应保持简洁、可发现，并从本入口文件可追踪。
- 对未来 Agent 可能破坏的重要行为，应补充可执行验证。
- 如果同类失败重复出现，应更新文档、脚本、测试或规则，让后续工作受到更明确的约束。
- 不得复制起点读书的品牌、UI 资产、精确布局或专有文本。

## 团队协作模型

本项目采用 Harness-Driven Agent Team 模型：

1. 用户提交需求或 Bug。
2. 主 Agent 识别任务类型、风险和影响范围。
3. 主 Agent 调用合适的 skill 进行分析。
4. 主 Agent 生成或更新 Harness 草案。
5. 用户确认 Harness 草案。
6. 主 Agent 按确认后的 Harness 拆分角色任务。
7. 产品、UI、前端、后端、测试等角色 Agent 在各自边界内执行。
8. 主 Agent 集成、验证，并把重复问题回填到 Harness。
