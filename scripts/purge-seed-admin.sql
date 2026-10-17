-- 清理种子账号 admin@cogniforge.local，并为 uk_users_name / uk_users_email 腾位。
-- 与 Go 启动时 PrepareUserNameUniqueness / PurgeSeedLocalAdmin 对齐。
-- 在能访问 cogniforge 库的机器上执行；先预览再 DELETE。

-- 1) 预览：种子邮箱（含软删）
SELECT id, email, name, role, status, deleted_at
FROM users
WHERE email = 'admin@cogniforge.local';

-- 2) 预览：活跃同名（部署 uk_users_name 前应只剩 1 条/名）
SELECT name, COUNT(*) AS cnt, array_agg(id ORDER BY created_at) AS ids
FROM users
WHERE deleted_at IS NULL
GROUP BY name
HAVING COUNT(*) > 1;

-- 3) 物理删除种子账号（含软删行）
-- BEGIN;
DELETE FROM users WHERE email = 'admin@cogniforge.local';
-- COMMIT;

-- 4)（可选）软删行若仍占原 name，改名释放唯一键
-- UPDATE users
-- SET name = left(name || '__del_' || left(id, 8) || '_' || extract(epoch from now())::bigint, 100)
-- WHERE deleted_at IS NOT NULL
--   AND name NOT LIKE '%__del_%';
