BEGIN;

DELETE FROM sys_user
WHERE username IN ('jack', 'test', 'vben')
  AND id NOT IN (
    'a0bb672a-a4b1-4ec9-807a-ba11e000d2a4',
    'e8a4dc57-a916-4d35-8b57-9f4dfae0d5b0',
    'f4f9e258-fa13-4467-95fb-c86019a377f9'
  );

INSERT INTO sys_user (id, create_time, update_time, username, password, nickname, email, status, avatar, "desc", extension, role_id)
VALUES
  ('a0bb672a-a4b1-4ec9-807a-ba11e000d2a4', '2025-02-26 18:59:32.728386+08', '2025-08-21 23:59:22.110609+08', 'jack', '123456', 'test1', '', 1, 'https://q1.qlogo.cn/g?b=qq&nk=190848757&s=640', '', '{"userRole":[{"role":"default"}],"email":"123456"}', 2),
  ('e8a4dc57-a916-4d35-8b57-9f4dfae0d5b0', '2025-08-22 17:51:43.539781+08', '2025-08-22 17:51:43.539782+08', 'test', 'testtest', '', '', 1, 'https://cdn.jsdelivr.net/gh/BaiMo-zyc/baimo.images@master/img/user-mini.png', '', '{"userRole":[{"role":"default"}],"email":"test"}', 0),
  ('f4f9e258-fa13-4467-95fb-c86019a377f9', '2023-05-17 22:29:18.185161+08', '2025-08-21 22:54:39.196075+08', 'vben', '123456', 'admin', '', 1, 'https://q1.qlogo.cn/g?b=qq&nk=190848757&s=640', 'test', '{"userRole":[{"role":"default"}],"email":"123456"}', 1)
ON CONFLICT (id) DO UPDATE SET
  create_time = EXCLUDED.create_time,
  update_time = EXCLUDED.update_time,
  username = EXCLUDED.username,
  password = EXCLUDED.password,
  nickname = EXCLUDED.nickname,
  email = EXCLUDED.email,
  status = EXCLUDED.status,
  avatar = EXCLUDED.avatar,
  "desc" = EXCLUDED."desc",
  extension = EXCLUDED.extension,
  role_id = EXCLUDED.role_id;

COMMIT;
