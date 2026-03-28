BEGIN;

INSERT INTO sys_dept (id, create_time, update_time, name, sort, status, "desc", extension, dom, pid)
VALUES
  (8, '2025-07-19 18:50:12.54752+08', '2025-07-19 18:50:12.547521+08', 'test', 3, true, '', '', 0, NULL),
  (9, '2025-07-19 18:50:16.996763+08', '2026-01-07 20:55:19.212511+08', 'test', 0, false, '', '', 0, NULL),
  (15, '2025-07-19 19:48:38.083376+08', '2026-01-07 20:54:57.34244+08', 'test3', 0, true, '', '', 0, 8)
ON CONFLICT (id) DO UPDATE SET
  create_time = EXCLUDED.create_time,
  update_time = EXCLUDED.update_time,
  name = EXCLUDED.name,
  sort = EXCLUDED.sort,
  status = EXCLUDED.status,
  "desc" = EXCLUDED."desc",
  extension = EXCLUDED.extension,
  dom = EXCLUDED.dom,
  pid = EXCLUDED.pid;

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
  (19, '2025-05-30 11:34:53.203+08', '2025-05-30 11:34:53.203+08', 16, 'menu', true, '/system/menu', '', '', 'Menu', '/system/menu/list', 'mdi:cloud-key-outline', 'system.menu.title', 1003, false, false, false, '', '', '', '/system/menu', 0, false, true, '', false, 0, false, false, false, false, true, '', '', '')
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

SELECT setval(pg_get_serial_sequence('sys_dept', 'id'), GREATEST(COALESCE((SELECT MAX(id) FROM sys_dept), 0), 15), true);
SELECT setval(pg_get_serial_sequence('sys_menu', 'id'), GREATEST(COALESCE((SELECT MAX(id) FROM sys_menu), 0), 19), true);

COMMIT;
