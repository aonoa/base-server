BEGIN;

DELETE FROM sys_user
WHERE username IN ('jack', 'vben')
  AND id NOT IN (
    'a0bb672a-a4b1-4ec9-807a-ba11e000d2a4',
    'f4f9e258-fa13-4467-95fb-c86019a377f9'
  );

INSERT INTO sys_user (id, create_time, update_time, username, password, nickname, email, status, avatar, "desc", extension, role_id)
VALUES
  ('a0bb672a-a4b1-4ec9-807a-ba11e000d2a4', '2025-02-26 18:59:32+08', '2025-08-21 23:59:22+08', 'jack', '123456', 'test1', '123456@example.com', 1, 'https://q1.qlogo.cn/g?b=qq&nk=190848757&s=640', '', '{"email":"123456@example.com"}', 2),
  ('f4f9e258-fa13-4467-95fb-c86019a377f9', '2023-05-17 22:29:18+08', '2025-08-21 22:54:39+08', 'vben', '123456', 'admin', 'admin@example.com', 1, 'https://q1.qlogo.cn/g?b=qq&nk=190848757&s=640', 'test', '{"email":"admin@example.com"}', 1)
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
