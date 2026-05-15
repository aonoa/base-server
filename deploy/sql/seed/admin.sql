BEGIN;

-- Permission master data is owned by the admin service.
INSERT INTO sys_service_registry (
  id, create_time, update_time, service_code, service_name, http_prefix, grpc_service, status, projection_enabled, description
)
VALUES
  ('svc-admin', '2026-04-18 00:00:00+08', '2026-04-18 00:00:00+08', 'admin', '平台管理服务', '/admin-api', 'api.admin.service.v1.AdminService', true, true, '平台控制面，负责菜单、服务注册、API 目录和投影源状态观测'),
  ('svc-auth', '2026-04-18 00:00:00+08', '2026-04-18 00:00:00+08', 'auth', '统一认证服务', '/auth-api', 'api.auth.service.v1.AuthService', true, false, '负责登录、令牌刷新、访问码下发、鉴权校验和权限快照接收'),
  ('svc-user', '2026-04-18 00:00:00+08', '2026-04-18 00:00:00+08', 'user', '用户中心服务', '/user-api', 'api.user.service.v1.UserService', true, false, '负责账号认证、用户资料、密码维护等身份基础能力'),
  ('svc-common', '2026-04-18 00:00:00+08', '2026-04-18 00:00:00+08', 'common', '公共能力服务', '/common-api', 'api.common.service.v1.CommonService, api.common.service.v1.UploadService, api.common.service.v1.SSEService', true, false, '负责文件上传、SSE 推送、Copilot 会话等平台公共能力'),
  ('svc-gateway', '2026-04-18 00:00:00+08', '2026-04-18 00:00:00+08', 'gateway', '统一网关服务', '/', '', true, false, '统一入口网关，负责路由转发、JWT 校验、Casbin 鉴权、限流和系统日志')
ON CONFLICT (service_code) DO UPDATE SET
  update_time = EXCLUDED.update_time,
  service_name = EXCLUDED.service_name,
  http_prefix = EXCLUDED.http_prefix,
  grpc_service = EXCLUDED.grpc_service,
  status = EXCLUDED.status,
  projection_enabled = EXCLUDED.projection_enabled,
  description = EXCLUDED.description;

DELETE FROM sys_service_registry
WHERE service_code NOT IN ('admin', 'auth', 'common', 'gateway', 'user');

DELETE FROM sys_projection_source_status
WHERE source_service NOT IN ('admin');

INSERT INTO sys_role (id, create_time, update_time, name, value, organization_id, status, "desc", menus, data_scope, data_scope_dept_ids)
VALUES
  (0, '2025-02-25 00:00:39.255+08', '2026-05-09 00:00:00+08', '默认角色', 'default', '9f740c1b-0210-4e3a-858d-d128edea924d', true, '', '[10, 11, 12, 13, 14, 15, 24]', 'self', '[]'),
  (1, '2025-02-25 00:00:39.255+08', '2025-08-22 00:47:56.788297+08', '超级管理员', 'root', '9f740c1b-0210-4e3a-858d-d128edea924d', true, '在系统层就拥有全部权限，不用设置', 'null', 'all', '[]'),
  (2, '2025-02-25 00:00:39.255+08', '2026-05-13 00:00:00+08', '管理员', 'admin', '9f740c1b-0210-4e3a-858d-d128edea924d', true, '', '[10, 11, 12, 13, 14, 15, 16, 17, 18, 3, 20, 22, 23, 24, 25, 26]', 'all', '[]')
ON CONFLICT (id) DO UPDATE SET
  create_time = EXCLUDED.create_time,
  update_time = EXCLUDED.update_time,
  name = EXCLUDED.name,
  value = EXCLUDED.value,
  organization_id = EXCLUDED.organization_id,
  status = EXCLUDED.status,
  "desc" = EXCLUDED."desc",
  menus = EXCLUDED.menus,
  data_scope = EXCLUDED.data_scope,
  data_scope_dept_ids = EXCLUDED.data_scope_dept_ids;

INSERT INTO sys_resources (id, create_time, update_time, name, type, value, method, description)
VALUES
  ('0ad0ceb4-6019-4feb-8d9a-eb561465777a', '2025-08-06 22:11:26.553496+08', '2025-08-06 22:11:26.553496+08', '系统管理菜单组', 'api', 'menu', '(GET|POST|PUT|DELETE)', ''),
  ('33c2ca63-4b51-43c6-b432-075047e083d7', '2025-08-06 22:11:02.317492+08', '2025-08-06 22:11:02.317492+08', '系统管理角色组', 'api', 'role', '(GET|POST|PUT|DELETE)', ''),
  ('779480bf-da02-46a4-8d87-89efbf1827de', '2025-08-06 22:10:40.309395+08', '2025-08-06 22:10:40.309395+08', '系统管理用户组', 'api', 'user', '(GET|POST|PUT|DELETE)', ''),
  ('7d6b49f5-3ef5-41e1-a4e5-0ccb96da5a95', '2025-08-06 22:09:27.466794+08', '2025-08-06 22:09:27.466795+08', '系统管理资源组', 'api', 'data', '(GET|POST|PUT|DELETE)', ''),
  ('a0f9309e-d04a-42c8-9bf5-7dd8b1af6e36', '2025-08-06 21:29:31.728041+08', '2025-08-20 19:03:27.329457+08', '基础api组', 'api', 'default', '(GET|POST|PUT|DELETE)', ''),
  ('a3e1fca5-e7ab-41f2-b7a9-2ab46f3640ea', '2025-08-06 22:08:19.563908+08', '2025-08-06 22:08:19.563908+08', '系统管理api组', 'api', 'api', '(GET|POST|PUT|DELETE)', ''),
  ('d3213f61-23d8-4be3-a44a-63f49d8c6cec', '2025-08-06 22:09:52.787831+08', '2025-08-06 22:09:52.787831+08', '系统管理部门组', 'api', 'dept', '(GET|POST|PUT|DELETE)', ''),
  ('6e24d8e7-d0e4-4c41-a19d-64fc2f1cf9df', '2026-05-13 00:00:00+08', '2026-05-13 00:00:00+08', '组织管理接口权限', 'api', 'organization', '(GET|POST|PUT|DELETE)', '组织管理页接口权限'),
  ('94d41386-8054-423a-9de2-af80f424b424', '2026-05-15 00:00:00+08', '2026-05-15 00:00:00+08', '组织权限范围接口权限', 'api', 'organization_permission_scope', '(GET|PUT)', '组织可用权限范围配置接口权限'),
  ('9da89181-e2d3-4b7a-9860-a12badd4b415', '2026-05-09 00:00:00+08', '2026-05-09 00:00:00+08', '站内信管理接口权限', 'api', 'site_message_manage', '(GET|POST|DELETE)', '站内信管理页接口权限'),
  ('f1ea1c6e-b1d4-4845-b0f5-07e2ddeae705', '2025-08-06 22:13:24.978161+08', '2025-08-06 22:13:24.978161+08', 'admin接口操作权限', 'api', 'admin', '(GET|POST|PUT|DELETE)', '')
ON CONFLICT (id) DO UPDATE SET
  create_time = EXCLUDED.create_time,
  update_time = EXCLUDED.update_time,
  name = EXCLUDED.name,
  type = EXCLUDED.type,
  value = EXCLUDED.value,
  method = EXCLUDED.method,
  description = EXCLUDED.description;

INSERT INTO resource_roles (resource_id, role_id)
VALUES
  ('33c2ca63-4b51-43c6-b432-075047e083d7', 2),
  ('779480bf-da02-46a4-8d87-89efbf1827de', 2),
  ('a3e1fca5-e7ab-41f2-b7a9-2ab46f3640ea', 1),
  ('a3e1fca5-e7ab-41f2-b7a9-2ab46f3640ea', 2),
  ('7d6b49f5-3ef5-41e1-a4e5-0ccb96da5a95', 1),
  ('7d6b49f5-3ef5-41e1-a4e5-0ccb96da5a95', 2),
  ('6e24d8e7-d0e4-4c41-a19d-64fc2f1cf9df', 1),
  ('6e24d8e7-d0e4-4c41-a19d-64fc2f1cf9df', 2),
  ('94d41386-8054-423a-9de2-af80f424b424', 1),
  ('94d41386-8054-423a-9de2-af80f424b424', 2),
  ('9da89181-e2d3-4b7a-9860-a12badd4b415', 1),
  ('9da89181-e2d3-4b7a-9860-a12badd4b415', 2),
  ('f1ea1c6e-b1d4-4845-b0f5-07e2ddeae705', 1),
  ('f1ea1c6e-b1d4-4845-b0f5-07e2ddeae705', 2),
  ('a0f9309e-d04a-42c8-9bf5-7dd8b1af6e36', 0),
  ('a0f9309e-d04a-42c8-9bf5-7dd8b1af6e36', 2),
  ('d3213f61-23d8-4be3-a44a-63f49d8c6cec', 2)
ON CONFLICT DO NOTHING;

INSERT INTO sys_api_resources (id, create_time, update_time, description, path, method, module, module_description, resources_group)
VALUES
  ('api-auth-login', '2025-08-04 14:33:11+08', '2025-08-04 15:16:38+08', '登录', '/auth-api/v1/login', 'POST', 'auth', '认证服务', 'default'),
  ('api-auth-codes', '2025-08-04 14:58:43+08', '2025-08-04 15:17:08+08', '获取用户权限码', '/auth-api/v1/access-codes', 'GET', 'auth', '认证服务', 'default'),
  ('api-auth-logout', '2025-08-04 14:59:21+08', '2025-08-04 15:17:15+08', '退出登录', '/auth-api/v1/logout', 'POST', 'auth', '认证服务', 'default'),
  ('api-auth-refresh', '2025-08-04 15:02:19+08', '2025-08-04 15:17:30+08', '刷新 token', '/auth-api/v1/refresh', 'POST', 'auth', '认证服务', 'default'),
  ('api-auth-reload-policy', '2025-08-04 15:02:54+08', '2025-08-04 15:18:14+08', '刷新 Casbin 权限缓存', '/auth-api/v1/reload-policy', 'POST', 'auth', '认证服务', 'admin'),
  ('api-auth-role-list', '2025-08-04 15:11:06+08', '2025-08-04 15:21:52+08', '获取角色列表', '/auth-api/v1/roles', 'GET', 'auth', '认证服务', 'role'),
  ('api-auth-role-create', '2025-08-04 15:11:35+08', '2025-08-04 15:22:00+08', '新增角色', '/auth-api/v1/roles', 'POST', 'auth', '认证服务', 'role'),
  ('api-auth-role-update', '2025-08-04 15:12:04+08', '2025-08-04 15:22:06+08', '更新角色', '/auth-api/v1/roles/{id}', 'PUT', 'auth', '认证服务', 'role'),
  ('api-auth-role-delete', '2025-08-04 15:12:16+08', '2025-08-04 15:22:13+08', '删除角色', '/auth-api/v1/roles/{id}', 'DELETE', 'auth', '认证服务', 'role'),
  ('api-auth-api-list', '2025-08-04 15:14:54+08', '2025-08-04 15:22:47+08', '获取 API 列表', '/auth-api/v1/apis', 'GET', 'auth', '认证服务', 'api'),
  ('api-auth-api-create', '2025-08-04 15:15:16+08', '2025-08-04 15:22:54+08', '新增 API', '/auth-api/v1/apis', 'POST', 'auth', '认证服务', 'api'),
  ('api-auth-api-update', '2025-08-04 15:15:29+08', '2025-08-04 15:23:01+08', '更新 API', '/auth-api/v1/apis/{id}', 'PUT', 'auth', '认证服务', 'api'),
  ('api-auth-api-delete', '2025-08-04 15:15:48+08', '2025-08-04 15:23:08+08', '删除 API', '/auth-api/v1/apis/{id}', 'DELETE', 'auth', '认证服务', 'api'),
  ('api-auth-resource-list', '2025-08-10 23:41:45+08', '2025-08-10 23:43:03+08', '获取资源列表', '/auth-api/v1/resources', 'GET', 'auth', '认证服务', 'data'),
  ('api-auth-resource-create', '2025-08-10 23:41:45+08', '2025-08-10 23:43:03+08', '新增资源', '/auth-api/v1/resources', 'POST', 'auth', '认证服务', 'data'),
  ('api-auth-resource-update', '2025-08-10 23:41:45+08', '2025-08-10 23:43:03+08', '更新资源', '/auth-api/v1/resources/{id}', 'PUT', 'auth', '认证服务', 'data'),
  ('api-auth-resource-delete', '2025-08-10 23:41:45+08', '2025-08-10 23:43:03+08', '删除资源', '/auth-api/v1/resources/{id}', 'DELETE', 'auth', '认证服务', 'data'),
  ('api-user-info', '2025-08-04 14:58:14+08', '2025-08-04 15:17:01+08', '获取用户信息', '/user-api/v1/users/{user_id}', 'GET', 'user', '用户服务', 'default'),
  ('api-user-list', '2025-08-04 15:03:23+08', '2025-08-04 15:19:21+08', '获取用户列表', '/user-api/v1/users', 'GET', 'user', '用户服务', 'user'),
  ('api-user-create', '2025-08-04 15:03:59+08', '2025-08-04 15:18:41+08', '新增用户', '/user-api/v1/users', 'POST', 'user', '用户服务', 'user'),
  ('api-user-update', '2025-08-04 15:04:24+08', '2025-08-04 15:18:56+08', '更新用户', '/user-api/v1/users/{id}', 'PUT', 'user', '用户服务', 'user'),
  ('api-user-delete', '2025-08-04 15:05:01+08', '2025-08-04 15:19:37+08', '删除用户', '/user-api/v1/users/{id}', 'DELETE', 'user', '用户服务', 'user'),
  ('api-user-check', '2025-08-04 15:05:44+08', '2025-08-04 15:19:59+08', '用户存在检查', '/user-api/v1/users/check', 'GET', 'user', '用户服务', 'default'),
  ('api-user-password', '2025-08-04 15:13:33+08', '2025-08-04 15:13:33+08', '修改用户密码', '/user-api/v1/users/{user_id}/password', 'POST', 'user', '用户服务', 'default'),
  ('api-admin-menu-current', '2025-08-04 15:06:10+08', '2025-08-04 15:20:12+08', '获取当前用户菜单', '/admin-api/v1/menus/current', 'GET', 'admin', '管理服务', 'default'),
  ('api-admin-role-list', '2026-05-14 00:00:00+08', '2026-05-14 00:00:00+08', '获取角色列表', '/admin-api/v1/roles', 'GET', 'admin', '管理服务', 'role'),
  ('api-admin-role-create', '2026-05-14 00:00:00+08', '2026-05-14 00:00:00+08', '新增角色', '/admin-api/v1/roles', 'POST', 'admin', '管理服务', 'role'),
  ('api-admin-role-update', '2026-05-14 00:00:00+08', '2026-05-14 00:00:00+08', '更新角色', '/admin-api/v1/roles/{id}', 'PUT', 'admin', '管理服务', 'role'),
  ('api-admin-role-delete', '2026-05-14 00:00:00+08', '2026-05-14 00:00:00+08', '删除角色', '/admin-api/v1/roles/{id}', 'DELETE', 'admin', '管理服务', 'role'),
  ('api-admin-menu-list', '2025-08-04 15:06:14+08', '2025-08-04 15:20:17+08', '获取系统菜单列表', '/admin-api/v1/menus', 'GET', 'admin', '管理服务', 'menu'),
  ('api-admin-menu-create', '2025-08-04 15:06:39+08', '2025-08-04 15:20:24+08', '新增菜单', '/admin-api/v1/menus', 'POST', 'admin', '管理服务', 'menu'),
  ('api-admin-menu-update', '2025-08-04 15:06:59+08', '2025-08-04 15:20:33+08', '更新菜单', '/admin-api/v1/menus/{id}', 'PUT', 'admin', '管理服务', 'menu'),
  ('api-admin-menu-delete', '2025-08-04 15:07:17+08', '2025-08-04 15:20:44+08', '删除菜单', '/admin-api/v1/menus/{id}', 'DELETE', 'admin', '管理服务', 'menu'),
  ('api-admin-menu-name-exists', '2025-08-04 15:08:32+08', '2025-08-04 15:20:58+08', '菜单名重复检查', '/admin-api/v1/menus/name-exists', 'GET', 'admin', '管理服务', 'menu'),
  ('api-admin-menu-path-exists', '2025-08-04 15:09:00+08', '2025-08-04 15:21:08+08', '菜单路径重复检查', '/admin-api/v1/menus/path-exists', 'GET', 'admin', '管理服务', 'menu'),
  ('api-admin-dept-list', '2025-08-04 15:09:41+08', '2025-08-04 15:21:16+08', '获取部门列表', '/admin-api/v1/depts', 'GET', 'admin', '管理服务', 'dept'),
  ('api-admin-dept-create', '2025-08-04 15:10:09+08', '2025-08-04 15:21:24+08', '新增部门', '/admin-api/v1/depts', 'POST', 'admin', '管理服务', 'dept'),
  ('api-admin-dept-update', '2025-08-04 15:10:26+08', '2025-08-04 15:21:36+08', '更新部门', '/admin-api/v1/depts/{id}', 'PUT', 'admin', '管理服务', 'dept'),
  ('api-admin-dept-delete', '2025-08-04 15:10:42+08', '2025-08-04 15:21:44+08', '删除部门', '/admin-api/v1/depts/{id}', 'DELETE', 'admin', '管理服务', 'dept'),
  ('api-admin-user-role-binding-list', '2026-04-30 00:00:00+08', '2026-04-30 00:00:00+08', '获取用户角色绑定列表', '/admin-api/v1/user-role-bindings', 'GET', 'admin', '管理服务', 'user'),
  ('api-admin-user-role-binding-get', '2026-04-30 00:00:00+08', '2026-04-30 00:00:00+08', '获取用户角色绑定', '/admin-api/v1/user-role-bindings/{user_id}', 'GET', 'admin', '管理服务', 'user'),
  ('api-admin-user-role-binding-upsert', '2026-04-30 00:00:00+08', '2026-04-30 00:00:00+08', '保存用户角色绑定', '/admin-api/v1/user-role-bindings/{user_id}', 'PUT', 'admin', '管理服务', 'user'),
  ('api-admin-user-role-binding-delete', '2026-04-30 00:00:00+08', '2026-04-30 00:00:00+08', '删除用户角色绑定', '/admin-api/v1/user-role-bindings/{user_id}', 'DELETE', 'admin', '管理服务', 'user'),
  ('api-admin-log-list', '2025-08-04 15:16:30+08', '2025-08-04 15:16:30+08', '获取日志列表', '/admin-api/v1/logs', 'GET', 'admin', '管理服务', 'admin'),
  ('api-admin-log-detail', '2025-08-04 15:16:35+08', '2025-08-04 15:16:35+08', '获取日志详情', '/admin-api/v1/logs/{id}', 'GET', 'admin', '管理服务', 'admin'),
  ('api-admin-api-list', '2026-04-30 00:00:00+08', '2026-04-30 00:00:00+08', '获取 API 列表', '/admin-api/v1/apis', 'GET', 'admin', '管理服务', 'api'),
  ('api-admin-api-create', '2026-04-30 00:00:00+08', '2026-04-30 00:00:00+08', '新增 API', '/admin-api/v1/apis', 'POST', 'admin', '管理服务', 'api'),
  ('api-admin-api-update', '2026-04-30 00:00:00+08', '2026-04-30 00:00:00+08', '更新 API', '/admin-api/v1/apis/{id}', 'PUT', 'admin', '管理服务', 'api'),
  ('api-admin-api-delete', '2026-04-30 00:00:00+08', '2026-04-30 00:00:00+08', '删除 API', '/admin-api/v1/apis/{id}', 'DELETE', 'admin', '管理服务', 'api'),
  ('api-admin-resource-list', '2026-04-30 00:00:00+08', '2026-04-30 00:00:00+08', '获取资源列表', '/admin-api/v1/resources', 'GET', 'admin', '管理服务', 'data'),
  ('api-admin-resource-create', '2026-04-30 00:00:00+08', '2026-04-30 00:00:00+08', '新增资源', '/admin-api/v1/resources', 'POST', 'admin', '管理服务', 'data'),
  ('api-admin-resource-update', '2026-04-30 00:00:00+08', '2026-04-30 00:00:00+08', '更新资源', '/admin-api/v1/resources/{id}', 'PUT', 'admin', '管理服务', 'data'),
  ('api-admin-resource-delete', '2026-04-30 00:00:00+08', '2026-04-30 00:00:00+08', '删除资源', '/admin-api/v1/resources/{id}', 'DELETE', 'admin', '管理服务', 'data'),
  ('api-admin-platform-service-list', '2026-04-30 00:00:00+08', '2026-04-30 00:00:00+08', '获取服务注册列表', '/admin-api/v1/platform/services', 'GET', 'admin', '管理服务', 'admin'),
  ('api-admin-platform-projection-source-list', '2026-04-30 00:00:00+08', '2026-04-30 00:00:00+08', '获取投影源状态列表', '/admin-api/v1/platform/projection-sources', 'GET', 'admin', '管理服务', 'admin'),
  ('api-admin-organization-list', '2026-05-13 00:00:00+08', '2026-05-13 00:00:00+08', '获取组织列表', '/admin-api/v1/organizations', 'GET', 'admin', '管理服务', 'organization'),
  ('api-admin-organization-create', '2026-05-13 00:00:00+08', '2026-05-13 00:00:00+08', '新增组织', '/admin-api/v1/organizations', 'POST', 'admin', '管理服务', 'organization'),
  ('api-admin-organization-update', '2026-05-13 00:00:00+08', '2026-05-13 00:00:00+08', '更新组织', '/admin-api/v1/organizations/{id}', 'PUT', 'admin', '管理服务', 'organization'),
  ('api-admin-organization-delete', '2026-05-13 00:00:00+08', '2026-05-13 00:00:00+08', '删除组织', '/admin-api/v1/organizations/{id}', 'DELETE', 'admin', '管理服务', 'organization'),
  ('api-admin-organization-member-list', '2026-05-13 00:00:00+08', '2026-05-13 00:00:00+08', '获取组织成员', '/admin-api/v1/organizations/{organization_id}/members', 'GET', 'admin', '管理服务', 'organization'),
  ('api-admin-organization-member-save', '2026-05-13 00:00:00+08', '2026-05-13 00:00:00+08', '保存组织成员', '/admin-api/v1/organizations/{organization_id}/members', 'PUT', 'admin', '管理服务', 'organization'),
  ('api-admin-organization-permission-scope-get', '2026-05-15 00:00:00+08', '2026-05-15 00:00:00+08', '获取组织权限范围', '/admin-api/v1/organizations/{organization_id}/permission-scope', 'GET', 'admin', '管理服务', 'organization_permission_scope'),
  ('api-admin-organization-permission-scope-save', '2026-05-15 00:00:00+08', '2026-05-15 00:00:00+08', '保存组织权限范围', '/admin-api/v1/organizations/{organization_id}/permission-scope', 'PUT', 'admin', '管理服务', 'organization_permission_scope'),
  ('api-admin-permission-catalog-current', '2026-05-15 00:00:00+08', '2026-05-15 00:00:00+08', '获取当前组织权限目录', '/admin-api/v1/permission-catalog/current', 'GET', 'admin', '管理服务', 'role'),
  ('api-admin-organization-permission-catalog', '2026-05-15 00:00:00+08', '2026-05-15 00:00:00+08', '获取指定组织权限目录', '/admin-api/v1/organizations/{organization_id}/permission-catalog', 'GET', 'admin', '管理服务', 'organization_permission_scope'),
  ('api-admin-my-organization-list', '2026-05-13 00:00:00+08', '2026-05-13 00:00:00+08', '获取我的组织', '/admin-api/v1/my/organizations', 'GET', 'admin', '管理服务', 'default'),
  ('api-admin-my-current-organization-switch', '2026-05-13 00:00:00+08', '2026-05-13 00:00:00+08', '切换当前组织', '/admin-api/v1/my/current-organization', 'PUT', 'admin', '管理服务', 'default'),
  ('api-admin-walk-route', '2026-03-28 00:00:00+08', '2026-03-28 00:00:00+08', '获取系统所有api接口', '/admin-api/v1/walk-routes', 'GET', 'admin', '系统管理', 'api'),
  ('api-common-site-message-my-list', '2026-05-09 00:00:00+08', '2026-05-09 00:00:00+08', '获取我的站内信列表', '/common-api/v1/site-messages/my', 'GET', 'common', '公共服务', 'default'),
  ('api-common-site-message-my-unread-count', '2026-05-09 00:00:00+08', '2026-05-09 00:00:00+08', '获取我的站内信未读数量', '/common-api/v1/site-messages/my/unread-count', 'GET', 'common', '公共服务', 'default'),
  ('api-common-site-message-my-read', '2026-05-09 00:00:00+08', '2026-05-09 00:00:00+08', '标记站内信已读', '/common-api/v1/site-messages/my/{message_id}/read', 'POST', 'common', '公共服务', 'default'),
  ('api-common-site-message-my-unread', '2026-05-09 00:00:00+08', '2026-05-09 00:00:00+08', '标记站内信未读', '/common-api/v1/site-messages/my/{message_id}/unread', 'POST', 'common', '公共服务', 'default'),
  ('api-common-site-message-my-read-all', '2026-05-09 00:00:00+08', '2026-05-09 00:00:00+08', '全部站内信标记已读', '/common-api/v1/site-messages/my/read-all', 'POST', 'common', '公共服务', 'default'),
  ('api-common-site-message-manage-list', '2026-05-09 00:00:00+08', '2026-05-09 00:00:00+08', '获取站内信发布记录', '/common-api/v1/site-messages/manage', 'GET', 'common', '公共服务', 'site_message_manage'),
  ('api-common-site-message-manage-create', '2026-05-09 00:00:00+08', '2026-05-09 00:00:00+08', '保存或发布站内信', '/common-api/v1/site-messages/manage', 'POST', 'common', '公共服务', 'site_message_manage'),
  ('api-common-site-message-manage-recall', '2026-05-09 00:00:00+08', '2026-05-09 00:00:00+08', '撤回站内信', '/common-api/v1/site-messages/manage/{id}/recall', 'POST', 'common', '公共服务', 'site_message_manage'),
  ('api-common-site-message-manage-delete', '2026-05-09 00:00:00+08', '2026-05-09 00:00:00+08', '删除未发布站内信', '/common-api/v1/site-messages/manage/{id}', 'DELETE', 'common', '公共服务', 'site_message_manage'),
  ('api-common-upload', '2025-08-04 15:16:13+08', '2025-08-04 15:23:14+08', '上传文件', '/common-api/v1/file/upload', 'POST', 'common', '公共服务', 'demo'),
  ('api-common-copilot-sse', '2025-08-04 15:16:20+08', '2025-08-04 15:23:20+08', 'Copilot SSE', '/common-api/v1/copilot/sse', 'POST', 'common', '公共服务', 'demo')
ON CONFLICT (path, method) DO UPDATE SET
  create_time = EXCLUDED.create_time,
  update_time = EXCLUDED.update_time,
  description = EXCLUDED.description,
  module = EXCLUDED.module,
  module_description = EXCLUDED.module_description,
  resources_group = EXCLUDED.resources_group;

INSERT INTO api_resources_roles (api_resources_id, role_id)
SELECT id, 1 FROM sys_api_resources
ON CONFLICT DO NOTHING;

INSERT INTO api_resources_roles (api_resources_id, role_id)
SELECT id, 2
FROM sys_api_resources
WHERE (path, method) IN (
  ('/admin-api/v1/roles', 'GET'),
  ('/admin-api/v1/roles', 'POST'),
  ('/admin-api/v1/roles/{id}', 'PUT'),
  ('/admin-api/v1/roles/{id}', 'DELETE'),
  ('/admin-api/v1/apis', 'GET'),
  ('/admin-api/v1/apis', 'POST'),
  ('/admin-api/v1/apis/{id}', 'PUT'),
  ('/admin-api/v1/apis/{id}', 'DELETE'),
  ('/admin-api/v1/resources', 'GET'),
  ('/admin-api/v1/resources', 'POST'),
  ('/admin-api/v1/resources/{id}', 'PUT'),
  ('/admin-api/v1/resources/{id}', 'DELETE'),
  ('/admin-api/v1/platform/services', 'GET'),
  ('/admin-api/v1/platform/projection-sources', 'GET'),
  ('/admin-api/v1/organizations', 'GET'),
  ('/admin-api/v1/organizations', 'POST'),
  ('/admin-api/v1/organizations/{id}', 'PUT'),
  ('/admin-api/v1/organizations/{id}', 'DELETE'),
  ('/admin-api/v1/organizations/{organization_id}/members', 'GET'),
  ('/admin-api/v1/organizations/{organization_id}/members', 'PUT'),
  ('/admin-api/v1/organizations/{organization_id}/permission-scope', 'GET'),
  ('/admin-api/v1/organizations/{organization_id}/permission-scope', 'PUT'),
  ('/admin-api/v1/permission-catalog/current', 'GET'),
  ('/admin-api/v1/organizations/{organization_id}/permission-catalog', 'GET'),
  ('/admin-api/v1/walk-routes', 'GET'),
  ('/common-api/v1/site-messages/manage', 'GET'),
  ('/common-api/v1/site-messages/manage', 'POST'),
  ('/common-api/v1/site-messages/manage/{id}/recall', 'POST'),
  ('/common-api/v1/site-messages/manage/{id}', 'DELETE')
)
ON CONFLICT DO NOTHING;

INSERT INTO sys_dept (id, create_time, update_time, name, sort, status, "desc", extension, pid, organization_id)
VALUES
  (8, '2025-07-19 18:50:12.54752+08', '2025-07-19 18:50:12.547521+08', 'test', 3, true, '', '', NULL, '9f740c1b-0210-4e3a-858d-d128edea924d'),
  (9, '2025-07-19 18:50:16.996763+08', '2026-01-07 20:55:19.212511+08', 'test', 0, false, '', '', NULL, '9f740c1b-0210-4e3a-858d-d128edea924d'),
  (15, '2025-07-19 19:48:38.083376+08', '2026-01-07 20:54:57.34244+08', 'test3', 0, true, '', '', 8, '9f740c1b-0210-4e3a-858d-d128edea924d')
ON CONFLICT (id) DO UPDATE SET
  create_time = EXCLUDED.create_time,
  update_time = EXCLUDED.update_time,
  name = EXCLUDED.name,
  sort = EXCLUDED.sort,
  status = EXCLUDED.status,
  "desc" = EXCLUDED."desc",
  extension = EXCLUDED.extension,
  pid = EXCLUDED.pid,
  organization_id = EXCLUDED.organization_id;

INSERT INTO sys_organization (id, create_time, update_time, name, code, sort, status, "desc", extension)
VALUES
  ('9f740c1b-0210-4e3a-858d-d128edea924d', '2026-05-13 00:00:00+08', '2026-05-13 00:00:00+08', '默认组织', 'default', 0, true, '系统默认组织', '')
ON CONFLICT (id) DO UPDATE SET
  update_time = EXCLUDED.update_time,
  name = EXCLUDED.name,
  code = EXCLUDED.code,
  sort = EXCLUDED.sort,
  status = EXCLUDED.status,
  "desc" = EXCLUDED."desc",
  extension = EXCLUDED.extension;

INSERT INTO sys_user_organization (create_time, update_time, user_id, organization_id, is_primary, status)
VALUES
  ('2026-05-13 00:00:00+08', '2026-05-13 00:00:00+08', 'a0bb672a-a4b1-4ec9-807a-ba11e000d2a4', '9f740c1b-0210-4e3a-858d-d128edea924d', true, true),
  ('2026-05-13 00:00:00+08', '2026-05-13 00:00:00+08', 'e8a4dc57-a916-4d35-8b57-9f4dfae0d5b0', '9f740c1b-0210-4e3a-858d-d128edea924d', false, true),
  ('2026-05-13 00:00:00+08', '2026-05-13 00:00:00+08', 'f4f9e258-fa13-4467-95fb-c86019a377f9', '9f740c1b-0210-4e3a-858d-d128edea924d', true, true)
ON CONFLICT (user_id, organization_id) DO UPDATE SET
  update_time = EXCLUDED.update_time,
  status = EXCLUDED.status;

INSERT INTO sys_menu (
  id, create_time, update_time, pid, type, status, path, redirect, alias, name, component, icon, title, "order",
  open_in_new_window, no_basic_layout, menu_visible_with_forbidden, link, iframe_src, active_icon, active_path,
  max_num_of_open_tab, keepalive, ignore_access, authority, affix_tab, affix_tab_order, hide_in_menu, hide_in_tab,
  hide_in_breadcrumb, hide_children_in_menu, full_path_key, badge, badge_type, badge_variants
)
VALUES
  (1, '2025-08-10 23:41:45.835283+08', '2025-08-10 23:43:03.512252+08', 16, 'menu', true, '/system/resource', '', '', 'Resource', '/system/resource/index', 'carbon:cloud-data-ops', '资源管理', 0, false, false, false, '', '', '', '/system/resource', 0, false, false, '', false, 0, false, false, false, false, false, '', '', ''),
  (2, '2025-08-23 21:06:21.149013+08', '2025-08-23 21:50:57.667662+08', 6, 'menu', true, '/log/system', '', '', 'SystemLog', '/log/system', 'carbon:ibm-watson-knowledge-catalog', '系统日志', 0, false, false, false, '', '', '', '/log/system', 0, false, false, '', false, 0, false, false, false, false, false, '', '', ''),
  (3, '2025-07-19 12:43:21.844839+08', '2025-07-19 12:54:47.743685+08', 16, 'menu', true, '/system/dept', '', '', 'Dept', '/system/dept/list', 'mdi:cloud-key-outline', 'system.dept.title', 0, false, false, false, '', '', '', '/system/dept', 0, false, false, '', false, 0, false, false, false, false, false, '', '', ''),
  (4, '2025-07-20 19:16:11.675085+08', '2025-11-01 15:55:09.635085+08', 0, 'menu', true, '/test', '', '', 'test', '/_core/about/index', 'carbon:test-tool', 'test', 0, false, false, false, '', '', '', '/test', 0, false, false, '', false, 0, false, false, false, false, false, '', '', ''),
  (5, '2025-08-04 13:04:42.372769+08', '2025-08-04 13:10:24.566445+08', 16, 'menu', true, '/system/api', '', '', 'Api', '/system/api/index', 'mdi:cloud-key-outline', 'API资源管理', 0, false, false, false, '', '', '', '/system/api', 0, false, false, '', false, 0, false, false, false, false, false, '', '', ''),
  (6, '2025-08-23 21:44:37.732462+08', '2025-08-23 21:44:37.732463+08', 0, 'catalog', true, '/log', '', '', 'Log', '', 'carbon:catalog', '日志审计', 0, false, false, false, '', '', '', '', 0, false, false, '', false, 0, false, false, false, false, false, '', '', ''),
  (7, '2025-08-23 21:50:21.941124+08', '2025-08-23 21:50:31.682528+08', 6, 'menu', true, '/log/login', '', '', 'LoginLog', '/log/login', 'carbon:catalog-publish', '登录日志', 0, false, false, false, '', '', '', '/log/login', 0, false, false, '', false, 0, false, false, false, false, false, '', '', ''),
  (8, '2025-08-23 22:52:35.362062+08', '2025-08-24 00:37:14.263147+08', 6, 'menu', true, '/log/logInfo/:id', '', '', 'LogInfo', '/log/log_info', 'carbon:ibm-knowledge-catalog-premium', '日志详情', 0, false, false, false, '', '', '', '/log/logInfo/:id', 0, false, false, '', false, 0, true, false, false, false, false, '', '', ''),
  (9, '2025-11-02 00:30:14.691863+08', '2025-11-08 20:57:20.367871+08', 13, 'embedded', false, '/demos/flowgram', '', '', 'FlowGram', '/demos/flowgram/index', 'carbon:branch', 'demos.flowgram', 0, false, false, false, '', 'http://localhost:3000/', '', '/demos/flowgram', 0, true, false, '', false, 0, false, false, false, false, false, '', '', ''),
  (10, '2024-07-08 22:25:32.213+08', '2024-07-08 22:25:32.213+08', 0, 'catalog', true, '/dashboard', '/analytics', '', 'Dashboard', '', 'lucide:layout-dashboard', 'page.dashboard.title', 1, false, false, false, '', '', '', '/dashboard', 0, false, true, '', false, 0, false, false, false, false, true, '', '', ''),
  (11, '2024-07-08 22:25:32.213+08', '2024-07-08 22:25:32.213+08', 10, 'menu', true, '/analytics', '', '', 'Analysis', '/dashboard/analytics/index', 'lucide:area-chart', 'page.dashboard.analytics', 1, false, false, false, '', '', '', '/analytics', 0, false, true, '', false, 0, false, false, false, false, true, '', '', ''),
  (12, '2024-07-08 22:25:32.213+08', '2024-07-08 22:25:32.213+08', 10, 'menu', true, '/workspace', '', '', 'Workspace', '/dashboard/workspace/index', 'ion:grid-outline', 'page.dashboard.workspace', 2, false, false, false, '', '', '', '/workspace', 0, false, true, '', false, 0, false, false, false, false, true, '', '', ''),
  (13, '2024-07-08 22:25:32.213+08', '2024-07-08 22:25:32.213+08', 0, 'catalog', true, '/demos', '/demos/access', '', 'Demos', '', 'ic:baseline-view-in-ar', 'demos.title', 1000, false, false, false, '', '', '', '/demos', 0, false, true, '', false, 0, false, false, false, false, true, '', '', ''),
  (14, '2024-07-08 22:25:32.213+08', '2024-07-08 22:25:32.213+08', 13, 'menu', true, 'antd', '', '', 'Antd', '/demos/antd/index', 'mdi:cloud-key-outline', 'demos.antd', 1001, false, false, false, '', '', '', 'antd', 0, false, true, '', false, 0, false, false, false, false, true, '', '', ''),
  (15, '2024-07-08 22:25:32.213+08', '2024-07-08 22:25:32.213+08', 13, 'menu', true, 'upload', '', '', 'Upload', '/demos/upload/index', 'mdi:cloud-key-outline', 'demos.upload', 1002, false, false, false, '', '', '', 'upload', 0, false, true, '', false, 0, false, false, false, false, true, '', '', ''),
  (16, '2024-07-08 22:25:32.213+08', '2024-07-08 22:25:32.213+08', 0, 'catalog', true, '/system', '/system', '', 'System', '', 'ic:baseline-view-in-ar', 'system.title', 5, false, false, false, '', '', '', '/system', 0, false, true, '', false, 0, false, false, false, false, true, '', '', ''),
  (17, '2024-07-08 22:25:32.213+08', '2025-07-19 12:47:42.919731+08', 16, 'menu', true, '/system/user', '', '', 'User', '/system/user/index', 'mdi:cloud-key-outline', 'system.user', 1001, false, false, false, '', '', '', '/system/user', 0, false, true, '', false, 0, false, false, false, false, false, '', '', ''),
  (18, '2024-07-08 22:25:32.213+08', '2025-07-19 23:39:04.581868+08', 16, 'menu', true, '/system/role', '', '', 'Role', '/system/role/list', 'mdi:cloud-key-outline', 'system.role.title', 1002, false, false, false, '', '', '', '/system/role', 0, false, true, '', false, 0, false, false, false, false, false, '', '', ''),
  (19, '2025-05-30 11:34:53.203+08', '2025-05-30 11:34:53.203+08', 16, 'menu', true, '/system/menu', '', '', 'Menu', '/system/menu/list', 'mdi:cloud-key-outline', 'system.menu.title', 1003, false, false, false, '', '', '', '/system/menu', 0, false, true, '', false, 0, false, false, false, false, true, '', '', ''),
  (20, '2026-04-20 00:00:00+08', '2026-04-20 00:00:00+08', 16, 'catalog', true, '/system/platform', '/system/platform/service', '', 'Platform', '', 'carbon:cloud-service-management', '平台治理', 1004, false, false, false, '', '', '', '/system/platform', 0, false, true, '', false, 0, false, false, false, false, true, '', '', ''),
  (22, '2026-04-20 00:00:00+08', '2026-04-20 00:00:00+08', 20, 'menu', true, '/system/platform/service', '', '', 'ServiceRegistry', '/system/platform/service/index', 'carbon:container-services', '服务注册', 1006, false, false, false, '', '', '', '/system/platform/service', 0, false, true, '', false, 0, false, false, false, false, false, '', '', ''),
  (23, '2026-04-20 00:00:00+08', '2026-04-20 00:00:00+08', 20, 'menu', true, '/system/platform/projection-source', '', '', 'ProjectionSource', '/system/platform/projection-source/index', 'carbon:data-check', '投影源状态', 1007, false, false, false, '', '', '', '/system/platform/projection-source', 0, false, true, '', false, 0, false, false, false, false, false, '', '', ''),
  (24, '2026-05-09 00:00:00+08', '2026-05-09 00:00:00+08', 0, 'menu', true, '/messages', '', '', 'SiteMessageInbox', '/_core/messages/inbox', 'lucide:mail', '站内信收件箱', 9999, false, false, false, '', '', '', '/messages', 0, false, false, '', false, 0, true, false, false, false, true, '', 'normal', 'success'),
  (25, '2026-05-09 00:00:00+08', '2026-05-09 00:00:00+08', 16, 'menu', true, '/system/site-message', '', '', 'SiteMessageManage', '/_core/messages/manage', 'lucide:mail', '站内信管理', 1008, false, false, false, '', '', '', '/system/site-message', 0, false, false, '', false, 0, false, false, false, false, true, '', 'normal', 'success'),
  (26, '2026-05-13 00:00:00+08', '2026-05-13 00:00:00+08', 16, 'menu', true, '/system/organization', '', '', 'Organization', '/system/organization/index', 'lucide:building-2', '组织管理', 1009, false, false, false, '', '', '', '/system/organization', 0, false, false, '', false, 0, false, false, false, false, true, '', '', '')
ON CONFLICT (id) DO UPDATE SET
  create_time = EXCLUDED.create_time,
  update_time = EXCLUDED.update_time,
  pid = EXCLUDED.pid,
  type = EXCLUDED.type,
  status = EXCLUDED.status,
  path = EXCLUDED.path,
  redirect = EXCLUDED.redirect,
  alias = EXCLUDED.alias,
  name = EXCLUDED.name,
  component = EXCLUDED.component,
  icon = EXCLUDED.icon,
  title = EXCLUDED.title,
  "order" = EXCLUDED."order",
  open_in_new_window = EXCLUDED.open_in_new_window,
  no_basic_layout = EXCLUDED.no_basic_layout,
  menu_visible_with_forbidden = EXCLUDED.menu_visible_with_forbidden,
  link = EXCLUDED.link,
  iframe_src = EXCLUDED.iframe_src,
  active_icon = EXCLUDED.active_icon,
  active_path = EXCLUDED.active_path,
  max_num_of_open_tab = EXCLUDED.max_num_of_open_tab,
  keepalive = EXCLUDED.keepalive,
  ignore_access = EXCLUDED.ignore_access,
  authority = EXCLUDED.authority,
  affix_tab = EXCLUDED.affix_tab,
  affix_tab_order = EXCLUDED.affix_tab_order,
  hide_in_menu = EXCLUDED.hide_in_menu,
  hide_in_tab = EXCLUDED.hide_in_tab,
  hide_in_breadcrumb = EXCLUDED.hide_in_breadcrumb,
  hide_children_in_menu = EXCLUDED.hide_children_in_menu,
  full_path_key = EXCLUDED.full_path_key,
  badge = EXCLUDED.badge,
  badge_type = EXCLUDED.badge_type,
  badge_variants = EXCLUDED.badge_variants;

INSERT INTO sys_organization_permission_scope (
  create_time, update_time, organization_id, permission_type, permission_ref, created_by
)
SELECT NOW(), NOW(), '9f740c1b-0210-4e3a-858d-d128edea924d', 'menu', id::text, ''
FROM sys_menu
ON CONFLICT (organization_id, permission_type, permission_ref) DO NOTHING;

INSERT INTO sys_organization_permission_scope (
  create_time, update_time, organization_id, permission_type, permission_ref, created_by
)
SELECT NOW(), NOW(), '9f740c1b-0210-4e3a-858d-d128edea924d', 'resource', id, ''
FROM sys_resources
ON CONFLICT (organization_id, permission_type, permission_ref) DO NOTHING;

INSERT INTO sys_user_role_binding (create_time, update_time, user_id, role_id, organization_id)
VALUES
  ('2025-02-26 18:59:32.728386+08', '2025-08-21 23:59:22.110609+08', 'a0bb672a-a4b1-4ec9-807a-ba11e000d2a4', 2, '9f740c1b-0210-4e3a-858d-d128edea924d'),
  ('2025-08-22 17:51:43.539781+08', '2025-08-22 17:51:43.539782+08', 'e8a4dc57-a916-4d35-8b57-9f4dfae0d5b0', 0, '9f740c1b-0210-4e3a-858d-d128edea924d'),
  ('2023-05-17 22:29:18.185161+08', '2025-08-21 22:54:39.196075+08', 'f4f9e258-fa13-4467-95fb-c86019a377f9', 1, '9f740c1b-0210-4e3a-858d-d128edea924d')
ON CONFLICT (user_id, role_id, organization_id) DO UPDATE SET
  create_time = EXCLUDED.create_time,
  update_time = EXCLUDED.update_time;

SELECT setval(pg_get_serial_sequence('sys_role', 'id'), GREATEST(COALESCE((SELECT MAX(id) FROM sys_role), 0), 2), true);
SELECT setval(pg_get_serial_sequence('sys_user_organization', 'id'), GREATEST(COALESCE((SELECT MAX(id) FROM sys_user_organization), 0), 1), true);
SELECT setval(pg_get_serial_sequence('sys_dept', 'id'), GREATEST(COALESCE((SELECT MAX(id) FROM sys_dept), 0), 15), true);
SELECT setval(pg_get_serial_sequence('sys_menu', 'id'), GREATEST(COALESCE((SELECT MAX(id) FROM sys_menu), 0), 26), true);

COMMIT;
