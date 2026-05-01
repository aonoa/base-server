# API URL 迁移说明

本文用于前端从旧单体接口迁移到新的拆分服务接口时，对照调整请求 URL。

## 总体变化

旧版本接口统一挂在 `basic-api` 前缀下，例如：

- `/basic-api/auth/**`
- `/basic-api/system/**`
- `/basic-api/menu/**`
- `/basic-api/v1/**`

重构后按服务拆成 4 组接口前缀：

- 认证服务：`/auth-api/v1/**`
- 用户服务：`/user-api/v1/**`
- 管理服务：`/admin-api/v1/**`
- 公共服务：`/common-api/v1/**`

如果前端仍然通过统一网关访问，只需要把旧路径替换成下表中的新路径。
如果前端会直连服务，还需要同时切换对应服务地址和端口：

- auth: `:8020`
- user: `:8010`
- admin: `:8030`
- common: `:8040`

---

## 旧前缀到新前缀的归类

| 旧前缀 | 新前缀 | 归属服务 |
| --- | --- | --- |
| `/basic-api/auth/*` | `/auth-api/v1/*` | auth |
| `/basic-api/system/user/*` | `/user-api/v1/*` | user |
| `/basic-api/system/role/*` | `/auth-api/v1/roles*` | auth |
| `/basic-api/system/api/*` | `/auth-api/v1/apis*` | auth |
| `/basic-api/system/resource/*` | `/auth-api/v1/resources*` | auth |
| `/basic-api/system/dept/*` | `/admin-api/v1/depts*` | admin |
| `/basic-api/system/menu/*` | `/admin-api/v1/menus*` | admin |
| `/basic-api/system/log/*` | `/admin-api/v1/logs*` | admin |
| `/basic-api/menu/*` | `/admin-api/v1/menus/current` | admin |
| `/basic-api/v1/server/file/*` | `/common-api/v1/file/*` | common |
| `/basic-api/v1/copilot/*` | `/common-api/v1/copilot/*` | common |

---

## 详细 URL 对照

## 1. 认证 auth

| 功能 | 旧 URL | 新 URL | 方法 |
| --- | --- | --- | --- |
| 登录 | `/basic-api/auth/login` | `/auth-api/v1/login` | POST |
| 获取权限码 | `/basic-api/auth/codes` | `/auth-api/v1/access-codes` | GET |
| 登出 | `/basic-api/auth/logout` | `/auth-api/v1/logout` | POST |
| 刷新 token | `/basic-api/auth/refresh` | `/auth-api/v1/refresh` | POST |
| 重载权限策略 | `/basic-api/auth/reloadPolicy` | `/auth-api/v1/reload-policy` | POST |

### 角色管理

| 功能 | 旧 URL | 新 URL | 方法 |
| --- | --- | --- | --- |
| 角色列表 | `/basic-api/system/role/list` | `/auth-api/v1/roles` | GET |
| 新增角色 | `/basic-api/system/role` | `/auth-api/v1/roles` | POST |
| 修改角色 | `/basic-api/system/role/{id}` | `/auth-api/v1/roles/{id}` | PUT |
| 删除角色 | `/basic-api/system/role/{id}` | `/auth-api/v1/roles/{id}` | DELETE |

### API 资源管理

| 功能 | 旧 URL | 新 URL | 方法 |
| --- | --- | --- | --- |
| API 列表 | `/basic-api/system/api/list` | `/auth-api/v1/apis` | GET |
| 新增 API | `/basic-api/system/api` | `/auth-api/v1/apis` | POST |
| 修改 API | `/basic-api/system/api/{id}` | `/auth-api/v1/apis/{id}` | PUT |
| 删除 API | `/basic-api/system/api/{id}` | `/auth-api/v1/apis/{id}` | DELETE |

### Resource 管理

| 功能 | 旧 URL | 新 URL | 方法 |
| --- | --- | --- | --- |
| Resource 列表 | `/basic-api/system/resource/list` | `/auth-api/v1/resources` | GET |
| 新增 Resource | `/basic-api/system/resource` | `/auth-api/v1/resources` | POST |
| 修改 Resource | `/basic-api/system/resource/{id}` | `/auth-api/v1/resources/{id}` | PUT |
| 删除 Resource | `/basic-api/system/resource/{id}` | `/auth-api/v1/resources/{id}` | DELETE |

### 已移除 / 不再对前端开放

| 旧 URL | 说明 |
| --- | --- |
| `/basic-api/system/setRoleStatus` | 对应 `SetRoleStatus`，新版本未提供 HTTP 接口 |

---

## 2. 用户 user

| 功能 | 旧 URL | 新 URL | 方法 |
| --- | --- | --- | --- |
| 用户列表 | `/basic-api/system/user/list` | `/user-api/v1/users` | GET |
| 新增用户 | `/basic-api/system/user` | `/user-api/v1/users` | POST |
| 修改用户 | `/basic-api/system/user/{id}` | `/user-api/v1/users/{id}` | PUT |
| 删除用户 | `/basic-api/system/user/{id}` | `/user-api/v1/users/{id}` | DELETE |
| 检查用户是否存在 | `/basic-api/system/user/user-exists` | `/user-api/v1/users/check` | **POST → GET** |

### 需要特别注意的 breaking changes

| 功能 | 旧 URL | 新 URL | 说明 |
| --- | --- | --- | --- |
| 获取当前用户信息 | `/basic-api/user/info` | `/user-api/v1/users/{user_id}` | 旧接口不带路径参数；新接口需要传 `user_id` |
| 修改密码 | `/basic-api/system/changePassword` | `/user-api/v1/users/{user_id}/password` | 新接口把 `user_id` 放进 URL 路径 |

### 参数变化说明

#### 检查用户是否存在

旧接口：
- `POST /basic-api/system/user/user-exists`
- 参数从 body 传入

新接口：
- `GET /user-api/v1/users/check`
- 参数通过 query 传入，例如：

```text
/user-api/v1/users/check?id=xxx&username=yyy
```

#### 修改密码

新接口示例：

```text
POST /user-api/v1/users/{user_id}/password
```

body 仍然传：

```json
{
  "user_id": "...",
  "password_old": "...",
  "password_new": "..."
}
```

前端至少需要把路径改成带 `user_id` 的形式。

---

## 3. 管理 admin

### 菜单

| 功能 | 旧 URL | 新 URL | 方法 |
| --- | --- | --- | --- |
| 获取菜单列表 | `/basic-api/system/menu/list` | `/admin-api/v1/menus` | GET |
| 获取路由菜单列表 | `/basic-api/menu/all` | `/admin-api/v1/menus/current` | GET |
| 菜单名是否存在 | `/basic-api/system/menu/name-exists` | `/admin-api/v1/menus/name-exists` | GET |
| 菜单路径是否存在 | `/basic-api/system/menu/path-exists` | `/admin-api/v1/menus/path-exists` | GET |
| 创建菜单 | `/basic-api/system/menu` | `/admin-api/v1/menus` | POST |
| 更新菜单 | `/basic-api/system/menu/{id}` | `/admin-api/v1/menus/{id}` | PUT |
| 删除菜单 | `/basic-api/system/menu/{id}` | `/admin-api/v1/menus/{id}` | DELETE |

### 部门

| 功能 | 旧 URL | 新 URL | 方法 |
| --- | --- | --- | --- |
| 部门列表 | `/basic-api/system/dept/list` | `/admin-api/v1/depts` | GET |
| 新增部门 | `/basic-api/system/dept` | `/admin-api/v1/depts` | POST |
| 修改部门 | `/basic-api/system/dept/{id}` | `/admin-api/v1/depts/{id}` | PUT |
| 删除部门 | `/basic-api/system/dept/{id}` | `/admin-api/v1/depts/{id}` | DELETE |

### 系统日志

| 功能 | 旧 URL | 新 URL | 方法 |
| --- | --- | --- | --- |
| 日志列表 | `/basic-api/system/log/list` | `/admin-api/v1/logs` | GET |
| 日志详情 | `/basic-api/system/log/{id}` | `/admin-api/v1/logs/{id}` | GET |

### 路由枚举

| 功能 | 旧 URL | 新 URL | 方法 |
| --- | --- | --- | --- |
| 获取系统所有 API 接口 | `/basic-api/system/getWalkRoute` | `/admin-api/v1/walk-routes` | GET |

### 已移除 / 不再对前端开放

| 旧 URL | 说明 |
| --- | --- |
| `/basic-api/system/setRoleStatus` | 对应 `SetRoleStatus`，新版本未提供 HTTP 接口 |

---

## 4. 公共 common

| 功能 | 旧 URL | 新 URL | 方法 |
| --- | --- | --- | --- |
| 文件上传 | `/basic-api/v1/server/file/upload` | `/common-api/v1/file/upload` | POST |
| Copilot SSE | `/basic-api/v1/copilot/sse` | `/common-api/v1/copilot/sse` | POST |

---

## 前端改造建议

## 1. 统一按服务拆分 API 模块

建议把前端请求按服务拆分：

- `authApi`
- `userApi`
- `adminApi`
- `commonApi`

不要再继续把所有 URL 放在一个 `basic-api` 模块下。

## 2. 重点检查以下调用点

优先全局搜索这些旧前缀：

```text
/basic-api/auth/
/basic-api/system/user/
/basic-api/system/role/
/basic-api/system/api/
/basic-api/system/resource/
/basic-api/system/dept/
/basic-api/system/menu/
/basic-api/system/log/
/basic-api/menu/
/basic-api/v1/server/file/
/basic-api/v1/copilot/
/basic-api/user/info
/basic-api/system/changePassword
/basic-api/system/user/user-exists
/basic-api/system/getWalkRoute
/basic-api/system/setRoleStatus
```

## 3. 最容易漏改的点

- `/basic-api/user/info` 不是简单换前缀，已经变成 `/user-api/v1/users/{user_id}`
- `/basic-api/system/changePassword` 不是简单换前缀，已经变成 `/user-api/v1/users/{user_id}/password`
- `/basic-api/system/user/user-exists` 从 `POST` 改成了 `GET`
- `/basic-api/menu/all` 需要切到 admin 服务的 `/admin-api/v1/menus/current`
- `/basic-api/auth/reloadPolicy` 改成了 kebab-case：`/auth-api/v1/reload-policy`

---

## 来源

当前对照基于以下 proto：

### 新接口
- `api/protos/auth/service/v1/auth.proto`
- `api/protos/user/service/v1/user.proto`
- `api/protos/admin/service/v1/admin.proto`
- `api/protos/common/service/v1/upload.proto`
- `api/protos/common/service/v1/sse.proto`

### 旧接口
- `api/protos/base_api/v1/base.proto`（重构前版本）
- `api/protos/base_api/v1/upload.proto`（重构前版本）
- `api/protos/base_api/v1/sse.proto`（重构前版本）
