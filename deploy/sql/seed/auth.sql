BEGIN;

INSERT INTO sys_role (id, create_time, update_time, name, value, status, "desc", menus)
VALUES
  (0, '2025-02-25 00:00:39.255+08', '2025-11-01 16:31:19.080417+08', '默认角色', 'default', true, '', '[10, 11, 12, 13, 14, 15]'),
  (1, '2025-02-25 00:00:39.255+08', '2025-08-22 00:47:56.788297+08', '超级管理员', 'root', true, '在系统层就拥有全部权限，不用设置', 'null'),
  (2, '2025-02-25 00:00:39.255+08', '2025-08-22 00:48:34.778634+08', '管理员', 'admin', true, '', '[10, 11, 12, 13, 14, 15, 16, 17, 18, 3]')
ON CONFLICT (id) DO UPDATE SET
  create_time = EXCLUDED.create_time,
  update_time = EXCLUDED.update_time,
  name = EXCLUDED.name,
  value = EXCLUDED.value,
  status = EXCLUDED.status,
  "desc" = EXCLUDED."desc",
  menus = EXCLUDED.menus;

INSERT INTO sys_resources (id, create_time, update_time, name, type, value, method, description)
VALUES
  ('0ad0ceb4-6019-4feb-8d9a-eb561465777a', '2025-08-06 22:11:26.553496+08', '2025-08-06 22:11:26.553496+08', '系统管理菜单组', 'api', 'menu', '(GET|POST|PUT|DELETE)', ''),
  ('33c2ca63-4b51-43c6-b432-075047e083d7', '2025-08-06 22:11:02.317492+08', '2025-08-06 22:11:02.317492+08', '系统管理角色组', 'api', 'role', '(GET|POST|PUT|DELETE)', ''),
  ('779480bf-da02-46a4-8d87-89efbf1827de', '2025-08-06 22:10:40.309395+08', '2025-08-06 22:10:40.309395+08', '系统管理用户组', 'api', 'user', '(GET|POST|PUT|DELETE)', ''),
  ('7d6b49f5-3ef5-41e1-a4e5-0ccb96da5a95', '2025-08-06 22:09:27.466794+08', '2025-08-06 22:09:27.466795+08', '系统管理资源组', 'api', 'data', '(GET|POST|PUT|DELETE)', ''),
  ('a0f9309e-d04a-42c8-9bf5-7dd8b1af6e36', '2025-08-06 21:29:31.728041+08', '2025-08-20 19:03:27.329457+08', '基础api组', 'api', 'default', '(GET|POST|PUT|DELETE)', ''),
  ('a3e1fca5-e7ab-41f2-b7a9-2ab46f3640ea', '2025-08-06 22:08:19.563908+08', '2025-08-06 22:08:19.563908+08', '系统管理api组', 'api', 'api', '(GET|POST|PUT|DELETE)', ''),
  ('d3213f61-23d8-4be3-a44a-63f49d8c6cec', '2025-08-06 22:09:52.787831+08', '2025-08-06 22:09:52.787831+08', '系统管理部门组', 'api', 'dept', '(GET|POST|PUT|DELETE)', ''),
  ('f1ea1c6e-b1d4-4845-b0f5-07e2ddeae705', '2025-08-06 22:13:24.978161+08', '2025-08-06 22:13:24.978161+08', 'admin接口操作权限', 'api', 'admin', 'GET', '')
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
  ('api-admin-log-list', '2025-08-04 15:16:30+08', '2025-08-04 15:16:30+08', '获取日志列表', '/admin-api/v1/logs', 'GET', 'admin', '管理服务', 'admin'),
  ('api-admin-log-detail', '2025-08-04 15:16:35+08', '2025-08-04 15:16:35+08', '获取日志详情', '/admin-api/v1/logs/{id}', 'GET', 'admin', '管理服务', 'admin'),
  ('api-common-upload', '2025-08-04 15:16:13+08', '2025-08-04 15:23:14+08', '上传文件', '/common-api/v1/file/upload', 'POST', 'common', '公共服务', 'demo'),
  ('api-common-copilot-sse', '2025-08-04 15:16:20+08', '2025-08-04 15:23:20+08', 'Copilot SSE', '/common-api/v1/copilot/sse', 'POST', 'common', '公共服务', 'demo')
ON CONFLICT (id) DO UPDATE SET
  create_time = EXCLUDED.create_time,
  update_time = EXCLUDED.update_time,
  description = EXCLUDED.description,
  path = EXCLUDED.path,
  method = EXCLUDED.method,
  module = EXCLUDED.module,
  module_description = EXCLUDED.module_description,
  resources_group = EXCLUDED.resources_group;

INSERT INTO api_resources_roles (api_resources_id, role_id)
SELECT id, 1 FROM sys_api_resources
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('sys_role', 'id'), GREATEST(COALESCE((SELECT MAX(id) FROM sys_role), 0), 2), true);

COMMIT;
