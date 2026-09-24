#!/usr/bin/env bash
# 清理监控请求日志：删掉 status_code >= 400 的失败记录，留下成功的。
# 在能访问 PostgreSQL 的机器上执行（本机或 SSH 到服务器）。
#
# 用法：
#   ./scripts/cleanup-error-request-logs.sh
#   PG_CONTAINER=pgsql ./scripts/cleanup-error-request-logs.sh
set -euo pipefail

PG_CONTAINER="${PG_CONTAINER:-pgsql}"
PG_USER="${PG_USER:-postgres}"
PG_DB="${PG_DB:-cogniforge}"

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

echo "=== 清理前 ==="
psql -c "
SELECT
    COUNT(*) FILTER (WHERE status_code < 400)  AS keep_success,
    COUNT(*) FILTER (WHERE status_code >= 400) AS delete_error,
    COUNT(*)                                   AS total
FROM request_logs;
"
psql -c "
SELECT status_code, COUNT(*) AS cnt
FROM request_logs
GROUP BY status_code
ORDER BY status_code;
"

read -r -p "确认删除失败记录（status_code >= 400），只留成功的？[y/N] " ans
if [[ "${ans}" != "y" && "${ans}" != "Y" ]]; then
  echo "已取消。"
  exit 0
fi

psql -c "DELETE FROM request_logs WHERE status_code >= 400;"

echo "=== 清理后 ==="
psql -c "
SELECT
    COUNT(*) FILTER (WHERE status_code < 400)  AS keep_success,
    COUNT(*) FILTER (WHERE status_code >= 400) AS delete_error,
    COUNT(*)                                   AS total
FROM request_logs;
"

echo "完成。刷新监控页即可。"
