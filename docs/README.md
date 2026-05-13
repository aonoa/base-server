# base-server Docs Index

当前仓库在 `monorepo` 分支上运行 `gateway + auth + user + admin + common` 五服务架构。下面这些文档是这条版本线的主要入口。

## 快速入口

- [README.md](../README.md)
  - 仓库级运行方式、代码生成、网关路由和本地调试说明
- [monorepo-overview.md](./monorepo-overview.md)
  - 当前微服务版本线的结构、服务边界、常见改动路径和已知缺口
- [permission-design.md](./permission-design.md)
  - 当前平台角色、菜单、API 资源组、服务前缀解析和 `scope_id` 权限设计
- [api-ownership.md](./api-ownership.md)
  - API 目录、服务归属和 `path + method` 唯一记录规则
- [table-ownership.md](./table-ownership.md)
  - 表归属、数据库归属和共享 Ent schema 的使用口径
- [auth-incremental-sync-design.md](./auth-incremental-sync-design.md)
  - `admin -> auth` 权限快照 / delta 投影链路

## 阅读顺序建议

1. 先读 [README.md](../README.md) 了解运行方式和入口目录。
2. 再读 [monorepo-overview.md](./monorepo-overview.md) 建立当前代码布局和服务职责的整体图。
3. 涉及权限、菜单、资源、API 目录时，继续读：
   - [permission-design.md](./permission-design.md)
   - [api-ownership.md](./api-ownership.md)
   - [auth-incremental-sync-design.md](./auth-incremental-sync-design.md)
4. 涉及数据库、Ent schema 或服务归库时，读 [table-ownership.md](./table-ownership.md)。
