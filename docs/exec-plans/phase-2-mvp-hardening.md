# Phase 2：MVP 验证和打磨

## 状态

本轮基础验证已完成，浏览器视觉验收可继续补充。

## 目标

在 Phase 1 已完成业务闭环的基础上，进行完整运行验证、端到端测试补强和体验打磨。

## 范围

包含：

- 启动 Docker daemon 后运行 `make docker-build`。
- 使用 `make dev` 启动完整服务栈。
- 运行 `make smoke` 验证注册、登录、搜索、阅读和管理员上传。
- 根据 smoke 结果修复真实运行问题。
- 补充更完整的前后端集成测试。
- 检查移动端和桌面端页面布局。

不包含：

- 新增非 MVP 功能。
- 推荐算法。
- 付费章节。
- 评论、书架、作者平台。

## 验收标准

- `make docker-build` 通过。
- `make dev` 的核心依赖路径已通过混合启动验证：Docker MySQL + 本地后端 + smoke。
- `make smoke` 通过。
- 前端关键页面已通过 TypeScript 构建；桌面和移动端截图验收待后续浏览器工具或人工验证。
- Harness 记录所有验证结果和残余风险。

## 本轮验证结果

- `make test`：通过。
- `make docker-build`：通过。
- `BASE_URL=http://localhost:8001 make smoke`：通过。
- 当前 `8000` 端口被本地后端进程占用，本轮使用 `8001` 启动当前代码验证真实 MySQL 链路。
- 浏览器视觉验收未自动执行，已记录为残余风险。
