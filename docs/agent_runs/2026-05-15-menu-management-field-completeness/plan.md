# 菜单管理字段完整性修复

## 范围

- 版本线：`base-server/monorepo` + `vben-admin/monorepo`。
- 变更级别：Level 1，本地缺陷修复。
- 目标：菜单管理创建、编辑、列表回显不再丢失后端已有字段，重点包括排序、标签页、徽章、访问控制和布局相关字段。

## 影响面

- 后端：`admin` 服务菜单 proto、菜单 Ent/Proto 转换逻辑，以及相关回归测试。
- 前端：`web-antd` 菜单管理抽屉表单、菜单 API 类型、中文文案。
- OpenAPI：后端 proto 新增字段后需要同步生成产物与前端 OpenAPI/类型。

## 验证计划

- 后端：`go test ./app/admin/service/internal/biz -count=1`。
- 后端生成：`make api`，必要时同步 OpenAPI。
- 前端：对菜单相关文件运行 eslint 或类型检查；若既有工程问题阻塞，记录具体错误。
- 全局：两仓库分别执行 `git diff --check`。

## 执行结果

- 已补齐 `Meta.fullPathKey`、`Meta.menuVisibleWithForbidden` proto 字段，并运行 `make api` 生成 Go/OpenAPI。
- 已补齐菜单 Ent/Proto 双向转换字段，包括排序、标签页、徽章、访问控制、布局和权限标识。
- 已增加菜单转换回归测试，覆盖字段保存、默认值和旧 `authCode` 入参兼容。
- 验证通过：`go test ./app/admin/service/internal/biz -count=1`。
- 验证通过：`git diff --check`。

## 注意事项

- 当前仓库已有上一轮站内信按组织推送的未提交改动，本轮不回滚、不重排这些改动。
- 菜单按钮类型使用现有 `button`，不再沿用旧的 `action` 判断。
- `authCode` 没有数据库列，本轮用 `meta.authority` 作为可持久化权限标识来源。
