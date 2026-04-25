package scheduler_test

import (
	"testing"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert"

	"cogniforge/internal/scheduler"
)

func TestScheduledWorkflow_Struct(t *testing.T) {
	now := time.Now()
	sw := scheduler.ScheduledWorkflow{
		ID:             "sched-1",
		WorkflowID:     "wf-1",
		UserID:         "user-1",
		CronExpression: "*/5 * * * *",
		Name:           "Test Schedule",
		IsActive:       true,
		LastRun:        &now,
		NextRun:        &now,
		RunCount:       10,
	}

	assert.Equal(t, "sched-1", sw.ID)
	assert.Equal(t, "wf-1", sw.WorkflowID)
	assert.Equal(t, "*/5 * * * *", sw.CronExpression)
	assert.True(t, sw.IsActive)
	assert.Equal(t, 10, sw.RunCount)
}

func TestCronExpression_Parsing(t *testing.T) {
	tests := []struct {
		name  string
		expr  string
		valid bool
	}{
		{"every minute", "* * * * *", true},
		{"every 5 minutes", "*/5 * * * *", true},
		{"every hour", "0 * * * *", true},
		{"every day at midnight", "0 0 * * *", true},
		{"every day at 3am", "0 3 * * *", true},
		{"every Monday at 9am", "0 9 * * 1", true},
		{"invalid expression", "invalid", false},
		{"empty expression", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := cron.ParseStandard(tt.expr)
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestNewScheduler(t *testing.T) {
	s := scheduler.NewScheduler()
	assert.NotNil(t, s)
}

func TestCronParser_WithSeconds(t *testing.T) {
	// Test that cron parser supports 5-field and 6-field expressions
	tests := []struct {
		name       string
		expr       string
		hasSeconds bool
	}{
		{"5-field standard", "* * * * *", false},
		{"6-field with seconds", "*/30 * * * * *", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var parser cron.Parser
			if tt.hasSeconds {
				parser = cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
			} else {
				parser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
			}

			schedule, err := parser.Parse(tt.expr)
			if tt.hasSeconds || tt.name == "5-field standard" {
				assert.NoError(t, err)
				assert.NotNil(t, schedule)
			}
		})
	}
}

func TestCronExpression_NextRun(t *testing.T) {
	tests := []struct {
		name string
		expr string
	}{
		{"every minute", "* * * * *"},
		{"every 5 minutes", "*/5 * * * *"},
		{"every hour", "0 * * * *"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
			schedule, err := parser.Parse(tt.expr)
			assert.NoError(t, err)

			now := time.Now()
			next := schedule.Next(now)

			// Next run should be after current time
			assert.True(t, next.After(now), "Next run should be after current time")
		})
	}
}

func TestScheduleState(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		isActive bool
		lastRun  *time.Time
		nextRun  *time.Time
	}{
		{"active schedule", true, &now, &now},
		{"inactive schedule", false, &now, nil},
		{"never run", true, nil, &now},
		{"paused", false, &now, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sw := scheduler.ScheduledWorkflow{
				IsActive: tt.isActive,
				LastRun:  tt.lastRun,
				NextRun:  tt.nextRun,
			}
			assert.Equal(t, tt.isActive, sw.IsActive)
		})
	}
}

func TestCronExpression_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name  string
		expr  string
		valid bool
	}{
		{"any value (*)", "* * * * *", true},
		{"specific value (5)", "5 * * * *", true},
		{"range (1-5)", "1-5 * * * *", true},
		{"step values (*/10)", "*/10 * * * *", true},
		{"list (1,3,5)", "1,3,5 * * * *", true},
		{"weekday list (mon,wed,fri)", "0 9 * * 1,3,5", true},
		{"weekday range (1-5)", "0 9 * * 1-5", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
			_, err := parser.Parse(tt.expr)
			if tt.valid {
				assert.NoError(t, err)
			}
		})
	}
}
