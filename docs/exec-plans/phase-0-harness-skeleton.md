# Phase 0：Harness 骨架

## 目标

在生成应用代码之前，先创建仓库内的初始 Harness 骨架。

## 范围

包含：

- 创建 `AGENTS.md`。
- 创建产品、架构、API、命令和 Harness 原则文档。
- 捕获初始决策。
- 将应用代码生成留到下一阶段。

不包含：

- 生成 Kratos 项目。
- 生成 React 项目。
- 生成 Dockerfile。
- 生成数据库 schema。
- 实现业务功能。

## 已确认输入

- 项目目录：`novel-reader`。
- 后端：Go + Kratos。
- 前端：Vite + React + TypeScript。
- 数据库：MySQL。
- 上传存储：本地文件系统。
- 上传格式：`.txt`。
- 初始功能：登录、注册、搜索、阅读、后台上传。
- 设计参考：只参考起点式阅读产品结构，不复制。

## 验收标准

- Harness 文档存在于项目目录下。
- 文档明确说明：只有在用户确认后才开始生成应用代码。
- 下一阶段可以使用这些文档作为上下文生成真实项目。

## 验证

本阶段使用手动验证：

```bash
find novel-reader -maxdepth 3 -type f | sort
```

Phase 0 不需要应用构建或测试命令。

## 状态

已完成。

