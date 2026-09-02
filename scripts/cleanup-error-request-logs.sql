-- 清理监控页失败请求，只留成功记录（status_code < 400）
-- 表：request_logs    错误定义与页面一致：status_code >= 400
-- 先看预览，确认后再跑 DELETE。

-- 1) 预览
SELECT
    COUNT(*) FILTER (WHERE status_code < 400)  AS keep_success,
    COUNT(*) FILTER (WHERE status_code >= 400) AS delete_error,
    COUNT(*)                                   AS total
FROM request_logs;

SELECT status_code, COUNT(*) AS cnt
FROM request_logs
GROUP BY status_code
ORDER BY status_code;

-- 2) 删除失败记录（确认预览无误后再执行）
BEGIN;
DELETE FROM request_logs WHERE status_code >= 400;

SELECT
    COUNT(*) FILTER (WHERE status_code < 400)  AS keep_success,
    COUNT(*) FILTER (WHERE status_code >= 400) AS delete_error,
    COUNT(*)                                   AS total
FROM request_logs;

-- 不对就 ROLLBACK;  对了再 COMMIT;
COMMIT;
