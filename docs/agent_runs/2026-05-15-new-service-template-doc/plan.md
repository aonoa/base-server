# 新增服务模板文档迭代记录

## 1. 目标

为 `base-server/monorepo` 微服务版本线补充“新增一个独立服务”的模板和步骤文档，降低后续新增服务时遗漏 proto、配置、Wire、Makefile、部署、gateway、权限和 OpenAPI 同步点的风险。

## 2. 变更级别

Level 1：文档更新。

本次不改变运行时代码、API 合约、数据库 schema、权限策略或前端行为。

## 3. 版本线

- 后端：`base-server/monorepo`
- 前端：`vben-admin/monorepo`

## 4. 影响范围

后端文档：

- `README.md`
- `docs/README.md`
- `docs/monorepo-overview.md`
- `docs/new-service-template.md`
- `docs/agent_runs/2026-05-15-new-service-template-doc/plan.md`

前端影响：

- 无。该任务只补后端新增服务步骤文档，没有 API 合约变化，不需要同步前端 OpenAPI 或修改页面。

## 5. 实施内容

- 新增 `docs/new-service-template.md`，覆盖新增服务的最小目录模板和执行步骤。
- 将新增服务步骤挂到后端根 README、docs README 和 monorepo overview。
- 明确本地配置与 Docker Compose 配置的地址差异。
- 明确新增服务时需要同步的权限/API 目录、服务注册、gateway、部署配置和前端 OpenAPI 边界。

## 6. 验证

已执行：

```bash
git diff --check
```

结果：通过。

未执行构建或测试：

- 本次为 docs-only 变更，没有 Go/Proto/Ent/前端代码改动。

## 7. 剩余风险

- 文档中的模板使用 `<service_code>`、`<ServiceName>` 等占位符；真正新增服务时仍需按实际服务名替换并运行对应生成命令。
- 新增真实服务时需要按当次服务职责再更新 `docs/api-ownership.md` 和 `docs/table-ownership.md`。

