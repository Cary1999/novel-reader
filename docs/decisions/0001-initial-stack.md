# 决策 0001：初始技术栈

## 状态

已接受，用于规划。

## 决策

MVP 使用以下技术栈：

- 后端：Go + Kratos
- 前端：Vite + React + TypeScript
- 数据库：MySQL
- 上传文件：本地文件系统
- 构建与运行：Docker 和 Docker Compose

## 理由

后端技术栈由需求明确指定。Vite + React + TypeScript 能让前端保持轻量，同时具备可维护性。MySQL 适合对用户、书籍、章节和上传记录进行熟悉的关系建模。本地文件系统让 MVP 的上传存储保持简单且易于检查。

