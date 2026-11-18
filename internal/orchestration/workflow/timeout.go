package workflow

import (
	"log/slog"
	"time"

	"cogniforge/internal/model"
)

// StartWaitTimeoutScanner 周期性将超时 waiting 标为 failed（阶段十八 W3）
func StartWaitTimeoutScanner(svc *WorkflowService, interval time.Duration) {
	if svc == nil {
		return
	}
	if interval <= 0 {
		interval = 30 * time.Second
	}
	go func() {
		// 启动时捞一把：进程重启后 waiting 仍在库中可恢复 / 可超时
		n := svc.CountWaiting()
		if n > 0 {
			slog.Info("workflow: recoverable waiting executions", "count", n)
		}
		_ = svc.FailTimedOutWaiting()

		t := time.NewTicker(interval)
		defer t.Stop()
		for range t.C {
			if err := svc.FailTimedOutWaiting(); err != nil {
				slog.Warn("workflow wait timeout scan failed", "error", err)
			}
		}
	}()
}

// CountWaiting waiting 数量（启动可观测）
func (s *WorkflowService) CountWaiting() int64 {
	var n int64
	_ = s.db.Model(&model.WorkflowExecution{}).Where("status = ?", "waiting").Count(&n).Error
	return n
}

// FailTimedOutWaiting 将 wait_until < now 的 waiting 置 failed
func (s *WorkflowService) FailTimedOutWaiting() error {
	now := time.Now()
	res := s.db.Model(&model.WorkflowExecution{}).
		Where("status = ? AND wait_until IS NOT NULL AND wait_until < ?", "waiting", now).
		Updates(map[string]any{
			"status":       "failed",
			"error":        "Wait timed out",
			"wait_reason":  "",
			"wait_until":   nil,
			"current_node": "",
			"completed_at": &now,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected > 0 {
		slog.Info("workflow: timed out waiting executions", "count", res.RowsAffected)
	}
	return nil
}
