package scheduler

import (
	"fmt"
	"log"
	"sync"
	"time"

	"cogniforge/internal/database"
	"cogniforge/internal/engine"
	"cogniforge/internal/model"
	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron        *cron.Cron
	engine      *engine.WorkflowEngine
	runningJobs map[string]cron.EntryID
	mu          sync.RWMutex
}

type ScheduledWorkflow struct {
	ID             string
	WorkflowID     string
	UserID         string
	CronExpression string
	Name           string
	IsActive       bool
	LastRun        *time.Time
	NextRun        *time.Time
	RunCount       int
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		cron:        cron.New(cron.WithSeconds()),
		engine:      engine.NewEngine(),
		runningJobs: make(map[string]cron.EntryID),
	}
}

func (s *Scheduler) Start() error {
	s.loadScheduledWorkflows()

	s.cron.Start()
	log.Println("Workflow scheduler started")
	return nil
}

func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
	log.Println("Workflow scheduler stopped")
}

func (s *Scheduler) loadScheduledWorkflows() {
	var schedules []model.WorkflowSchedule
	if err := database.DB.Where("is_active = ?", true).Find(&schedules).Error; err != nil {
		log.Printf("Failed to load scheduled workflows: %v", err)
		return
	}

	for _, schedule := range schedules {
		if err := s.AddSchedule(&schedule); err != nil {
			log.Printf("Failed to add schedule %s: %v", schedule.ID, err)
		}
	}
}

func (s *Scheduler) AddSchedule(schedule *model.WorkflowSchedule) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.runningJobs[schedule.ID]; exists {
		return fmt.Errorf("schedule %s already running", schedule.ID)
	}

	entryID, err := s.cron.AddFunc(schedule.CronExpression, func() {
		s.executeScheduledWorkflow(schedule.ID)
	})
	if err != nil {
		return fmt.Errorf("invalid cron expression: %w", err)
	}

	s.runningJobs[schedule.ID] = entryID
	log.Printf("Added schedule %s with cron %s", schedule.ID, schedule.CronExpression)
	return nil
}

func (s *Scheduler) RemoveSchedule(scheduleID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entryID, exists := s.runningJobs[scheduleID]
	if !exists {
		return fmt.Errorf("schedule %s not found", scheduleID)
	}

	s.cron.Remove(entryID)
	delete(s.runningJobs, scheduleID)
	log.Printf("Removed schedule %s", scheduleID)
	return nil
}

func (s *Scheduler) executeScheduledWorkflow(scheduleID string) {
	var schedule model.WorkflowSchedule
	if err := database.DB.Where("id = ?", scheduleID).First(&schedule).Error; err != nil {
		log.Printf("Schedule %s not found: %v", scheduleID, err)
		return
	}

	if !schedule.IsActive {
		log.Printf("Schedule %s is inactive, skipping", scheduleID)
		return
	}

	log.Printf("Executing scheduled workflow %s", schedule.WorkflowID)

	executionID, err := s.engine.ExecuteAsync(schedule.WorkflowID, schedule.UserID, schedule.DefaultInput)
	if err != nil {
		log.Printf("Failed to execute scheduled workflow %s: %v", schedule.WorkflowID, err)
		s.recordFailedRun(&schedule)
		return
	}

	now := time.Now()
	database.DB.Model(&schedule).Updates(map[string]any{
		"last_run":  &now,
		"run_count": schedule.RunCount + 1,
	})

	log.Printf("Scheduled workflow %s started with execution %s", schedule.WorkflowID, executionID)
}

func (s *Scheduler) recordFailedRun(schedule *model.WorkflowSchedule) {
	database.DB.Model(schedule).Update("last_error", fmt.Sprintf("Failed at %s", time.Now().Format(time.RFC3339)))
}

func (s *Scheduler) ListSchedules(userID string) ([]model.WorkflowSchedule, error) {
	var schedules []model.WorkflowSchedule
	query := database.DB.Model(&model.WorkflowSchedule{})

	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.Order("created_at DESC").Find(&schedules).Error; err != nil {
		return nil, err
	}

	for i := range schedules {
		s.mu.RLock()
		if entryID, exists := s.runningJobs[schedules[i].ID]; exists {
			if entry := s.cron.Entry(entryID); entry.ID != 0 {
				nextRun := entry.Next
				schedules[i].NextRun = &nextRun
			}
		}
		s.mu.RUnlock()
	}

	return schedules, nil
}

func (s *Scheduler) GetSchedule(scheduleID string) (*model.WorkflowSchedule, error) {
	var schedule model.WorkflowSchedule
	if err := database.DB.Where("id = ?", scheduleID).First(&schedule).Error; err != nil {
		return nil, err
	}

	s.mu.RLock()
	if entryID, exists := s.runningJobs[schedule.ID]; exists {
		if entry := s.cron.Entry(entryID); entry.ID != 0 {
			nextRun := entry.Next
			schedule.NextRun = &nextRun
		}
	}
	s.mu.RUnlock()

	return &schedule, nil
}

func (s *Scheduler) PauseSchedule(scheduleID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entryID, exists := s.runningJobs[scheduleID]
	if !exists {
		return fmt.Errorf("schedule %s not found", scheduleID)
	}

	s.cron.Remove(entryID)
	delete(s.runningJobs, scheduleID)

	database.DB.Model(&model.WorkflowSchedule{}).Where("id = ?", scheduleID).Update("is_active", false)
	log.Printf("Paused schedule %s", scheduleID)
	return nil
}

func (s *Scheduler) ResumeSchedule(scheduleID string) error {
	var schedule model.WorkflowSchedule
	if err := database.DB.Where("id = ?", scheduleID).First(&schedule).Error; err != nil {
		return err
	}

	schedule.IsActive = true
	if err := database.DB.Save(&schedule).Error; err != nil {
		return err
	}

	return s.AddSchedule(&schedule)
}

func (s *Scheduler) RunNow(scheduleID string) (string, error) {
	var schedule model.WorkflowSchedule
	if err := database.DB.Where("id = ?", scheduleID).First(&schedule).Error; err != nil {
		return "", err
	}

	return s.engine.ExecuteAsync(schedule.WorkflowID, schedule.UserID, schedule.DefaultInput)
}
