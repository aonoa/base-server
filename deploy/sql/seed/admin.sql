BEGIN;

INSERT INTO sys_dept (id, create_time, update_time, name, sort, status, "desc", extension, dom, pid)
VALUES
  (1, '2025-07-19 18:50:12+08', '2025-07-19 18:50:12+08', '总部', 0, true, '默认根部门', '', 0, NULL),
  (2, '2025-07-19 19:48:38+08', '2025-07-19 19:48:38+08', '研发部', 1, true, '默认研发部门', '', 0, 1)
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
  (1, '2025-08-10 23:41:45+08', '2025-08-10 23:43:03+08', 16, 'menu', true, '/system/resource', '', '', 'Resource', '/system/resource/index', 'carbon:cloud-data-ops', '资源管理', 0, false, false, false, '', '', '', '/system/resource', 0, false, false, '', false, 0, false, false, false, false, true, '', 'normal', 'success'),
  (3, '2025-07-19 12:43:21+08', '2025-07-19 12:54:47+08', 16, 'menu', true, '/system/dept', '', '', 'Dept', '/system/dept/list', 'mdi:cloud-key-outline', 'system.dept.title', 0, false, false, false, '', '', '', '/system/dept', 0, false, false, '', false, 0, false, false, false, false, true, '', 'normal', 'success'),
  (5, '2025-08-04 13:04:42+08', '2025-08-04 13:10:24+08', 16, 'menu', true, '/system/api', '', '', 'Api', '/system/api/index', 'mdi:cloud-key-outline', 'API资源管理', 0, false, false, false, '', '', '', '/system/api', 0, false, false, '', false, 0, false, false, false, false, true, '', 'normal', 'success'),
  (10, '2024-07-08 22:25:32+08', '2024-07-08 22:25:32+08', 0, 'catalog', true, '/dashboard', '/analytics', '', 'Dashboard', '', 'lucide:layout-dashboard', 'page.dashboard.title', 1, false, false, false, '', '', '', '/dashboard', 0, false, true, '', false, 0, false, false, false, false, true, '', 'normal', 'success'),
  (11, '2024-07-08 22:25:32+08', '2024-07-08 22:25:32+08', 10, 'menu', true, '/analytics', '', '', 'Analysis', '/dashboard/analytics/index', 'lucide:area-chart', 'page.dashboard.analytics', 1, false, false, false, '', '', '', '/analytics', 0, false, true, '', false, 0, false, false, false, false, true, '', 'normal', 'success'),
  (12, '2024-07-08 22:25:32+08', '2024-07-08 22:25:32+08', 10, 'menu', true, '/workspace', '', '', 'Workspace', '/dashboard/workspace/index', 'ion:grid-outline', 'page.dashboard.workspace', 2, false, false, false, '', '', '', '/workspace', 0, false, true, '', false, 0, false, false, false, false, true, '', 'normal', 'success'),
  (13, '2024-07-08 22:25:32+08', '2024-07-08 22:25:32+08', 0, 'catalog', true, '/demos', '/demos/access', '', 'Demos', '', 'ic:baseline-view-in-ar', 'demos.title', 1000, false, false, false, '', '', '', '/demos', 0, false, true, '', false, 0, false, false, false, false, true, '', 'normal', 'success'),
  (14, '2024-07-08 22:25:32+08', '2024-07-08 22:25:32+08', 13, 'menu', true, 'antd', '', '', 'Antd', '/demos/antd/index', 'mdi:cloud-key-outline', 'demos.antd', 1001, false, false, false, '', '', '', 'antd', 0, false, true, '', false, 0, false, false, false, false, true, '', 'normal', 'success'),
  (15, '2024-07-08 22:25:32+08', '2024-07-08 22:25:32+08', 13, 'menu', true, 'upload', '', '', 'Upload', '/demos/upload/index', 'mdi:cloud-key-outline', 'demos.upload', 1002, false, false, false, '', '', '', 'upload', 0, false, true, '', false, 0, false, false, false, false, true, '', 'normal', 'success'),
  (16, '2024-07-08 22:25:32+08', '2024-07-08 22:25:32+08', 0, 'catalog', true, '/system', '/system', '', 'System', '', 'ic:baseline-view-in-ar', 'system.title', 5, false, false, false, '', '', '', '/system', 0, false, true, '', false, 0, false, false, false, false, true, '', 'normal', 'success'),
  (17, '2025-07-19 12:47:42+08', '2025-07-19 12:47:42+08', 16, 'menu', true, '/system/user', '', '', 'User', '/system/user/index', 'mdi:cloud-key-outline', 'system.user', 1001, false, false, false, '', '', '', '/system/user', 0, false, true, '', false, 0, false, false, false, false, true, '', 'normal', 'success'),
  (18, '2025-07-19 23:39:04+08', '2025-07-19 23:39:04+08', 16, 'menu', true, '/system/role', '', '', 'Role', '/system/role/list', 'mdi:cloud-key-outline', 'system.role.title', 1002, false, false, false, '', '', '', '/system/role', 0, false, true, '', false, 0, false, false, false, false, true, '', 'normal', 'success'),
  (19, '2025-05-30 11:34:53+08', '2025-05-30 11:34:53+08', 16, 'menu', true, '/system/menu', '', '', 'Menu', '/system/menu/list', 'mdi:cloud-key-outline', 'system.menu.title', 1003, false, false, false, '', '', '', '/system/menu', 0, false, true, '', false, 0, false, false, false, false, true, '', 'normal', 'success')
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

SELECT setval(pg_get_serial_sequence('sys_dept', 'id'), GREATEST(COALESCE((SELECT MAX(id) FROM sys_dept), 0), 2), true);
SELECT setval(pg_get_serial_sequence('sys_menu', 'id'), GREATEST(COALESCE((SELECT MAX(id) FROM sys_menu), 0), 19), true);

COMMIT;
