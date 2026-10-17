#!/usr/bin/env bash
# 物理删除种子账号 admin@cogniforge.local（含软删行），避免与真实 admin 用户名冲突。
# 与 Go 启动时 auth.PurgeSeedLocalAdmin 对齐。在能访问 PostgreSQL 的机器上执行。
#
# 用法：
#   ./scripts/purge-seed-admin.sh
#   PG_CONTAINER=pgsql ./scripts/purge-seed-admin.sh
set -euo pipefail

PG_CONTAINER="${PG_CONTAINER:-pgsql}"
PG_USER="${PG_USER:-postgres}"
PG_DB="${PG_DB:-cogniforge}"
SEED_EMAIL="${SEED_EMAIL:-admin@cogniforge.local}"

if ! docker ps --format '{{.Names}}' | grep -qx "$PG_CONTAINER"; then
  echo "找不到容器 $PG_CONTAINER。可用容器："
  docker ps --format '  {{.Names}}'
  echo
  echo "改容器名再跑，例如：PG_CONTAINER=pgsql $0"
  exit 1
fi

psql() {
  docker exec -i "$PG_CONTAINER" psql -U "$PG_USER" -d "$PG_DB" -v ON_ERROR_STOP=1 "$@"
}

echo "=== 种子账号（含软删）==="
psql -c "
SELECT id, email, name, role, status, deleted_at
FROM users
WHERE email = '${SEED_EMAIL}';
"

echo "=== 活跃同名预览（部署唯一索引前应每名最多 1 条）==="
psql -c "
SELECT name, COUNT(*) AS cnt, array_agg(id ORDER BY created_at) AS ids
FROM users
WHERE deleted_at IS NULL
GROUP BY name
HAVING COUNT(*) > 1;
"

read -r -p "确认物理删除邮箱 ${SEED_EMAIL}（含软删行）？[y/N] " ans
if [[ "${ans}" != "y" && "${ans}" != "Y" ]]; then
  echo "已取消。"
  exit 0
fi

psql -c "DELETE FROM users WHERE email = '${SEED_EMAIL}';"

echo "=== 删除后 ==="
psql -c "
SELECT COUNT(*) AS remaining
FROM users
WHERE email = '${SEED_EMAIL}';
"

echo "完成。重启 Go 服务后会跑 PrepareUserNameUniqueness + AutoMigrate(uk_users_name)。"
