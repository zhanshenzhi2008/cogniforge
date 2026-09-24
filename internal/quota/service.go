package quota

import (
	"context"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"cogniforge/internal/model"
	"cogniforge/internal/response"
	"cogniforge/internal/trace"
)

const (
	DefaultDailyRequests = 30
	DefaultDailyTokens   = int64(100000)
	DefaultMonthlyTokens = int64(1000000)
	DefaultRPM           = 8

	statusOK      = "ok"
	statusBlocked = "quota_blocked"
)

var shanghai *time.Location

func init() {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*3600)
	}
	shanghai = loc
}

type Service struct {
	db    *gorm.DB
	store Store
}

func New(db *gorm.DB, store Store) *Service {
	return &Service{db: db, store: store}
}

func (s *Service) EnsureDefaultPolicy() {
	if s.db == nil {
		return
	}
	var n int64
	s.db.Model(&model.QuotaPolicy{}).Where("user_id IS NULL").Count(&n)
	if n > 0 {
		return
	}
	p := model.QuotaPolicy{
		ID:             newID(),
		DailyRequests:  DefaultDailyRequests,
		DailyTokens:    DefaultDailyTokens,
		MonthlyTokens:  DefaultMonthlyTokens,
		RPM:            DefaultRPM,
		AdminUnlimited: true,
		Enabled:        true,
	}
	if err := s.db.Create(&p).Error; err != nil {
		slog.Warn("quota default policy insert failed", "error", err)
	}
}

type Snapshot struct {
	Unlimited      bool   `json:"unlimited"`
	Day            Window `json:"day"`
	Month          Window `json:"month"`
	Warn           bool   `json:"warn"`
	AdminUnlimited bool   `json:"admin_unlimited"`
}

type Window struct {
	RequestsUsed  int64     `json:"requests_used,omitempty"`
	RequestsLimit int       `json:"requests_limit,omitempty"`
	TokensUsed    int64     `json:"tokens_used"`
	TokensLimit   int64     `json:"tokens_limit"`
	ResetsAt      time.Time `json:"resets_at"`
}

type resolvedPolicy struct {
	dailyReq  int
	dailyTok  int64
	monthTok  int64
	rpm       int
	unlimited bool
	enabled   bool
}

func (s *Service) resolvePolicy(userID, role string) resolvedPolicy {
	def := resolvedPolicy{
		dailyReq:  DefaultDailyRequests,
		dailyTok:  DefaultDailyTokens,
		monthTok:  DefaultMonthlyTokens,
		rpm:       DefaultRPM,
		unlimited: role == "admin",
		enabled:   true,
	}
	if s.db == nil {
		return def
	}
	var global model.QuotaPolicy
	if err := s.db.Where("user_id IS NULL").First(&global).Error; err == nil {
		def.dailyReq = global.DailyRequests
		def.dailyTok = global.DailyTokens
		def.monthTok = global.MonthlyTokens
		def.rpm = global.RPM
		def.unlimited = role == "admin" && global.AdminUnlimited
		def.enabled = global.Enabled
	}
	var override model.QuotaPolicy
	if err := s.db.Where("user_id = ?", userID).First(&override).Error; err == nil {
		def.dailyReq = override.DailyRequests
		def.dailyTok = override.DailyTokens
		def.monthTok = override.MonthlyTokens
		def.rpm = override.RPM
		def.enabled = override.Enabled
		if role == "admin" {
			def.unlimited = override.AdminUnlimited
		} else {
			def.unlimited = false
		}
	}
	return def
}

func nowSH() time.Time {
	return time.Now().In(shanghai)
}

func dayKey(userID string, t time.Time) (req, tokens string) {
	d := t.Format("20060102")
	return "cogniforge:quota:user:" + userID + ":day:" + d + ":req",
		"cogniforge:quota:user:" + userID + ":day:" + d + ":tokens"
}

func monthKey(userID string, t time.Time) string {
	return "cogniforge:quota:user:" + userID + ":month:" + t.Format("200601") + ":tokens"
}

func rpmKey(userID string, t time.Time) string {
	return "cogniforge:quota:rl:" + userID + ":" + t.Format("200601021504")
}

func nextMidnight(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d+1, 0, 0, 0, 0, t.Location())
}

func nextMonth(t time.Time) time.Time {
	y, m, _ := t.Date()
	return time.Date(y, m+1, 1, 0, 0, 0, 0, t.Location())
}

func (s *Service) loadUser(userID string) (role string, err error) {
	if s.db == nil {
		return "user", nil
	}
	var u model.User
	if err := s.db.Select("id", "role").Where("id = ?", userID).First(&u).Error; err != nil {
		return "", err
	}
	return u.Role, nil
}

// Allow 对话前检查。成功时已占用 1 次日请求。失败不打上游。
func (s *Service) Allow(ctx context.Context, userID, source string) error {
	if userID == "" {
		return nil
	}
	if s.store == nil {
		return ErrUnavailable
	}
	role, err := s.loadUser(userID)
	if err != nil {
		return err
	}
	p := s.resolvePolicy(userID, role)
	if !p.enabled {
		return ErrExceeded
	}
	if p.unlimited {
		return nil
	}

	now := nowSH()
	reqKey, tokKey := dayKey(userID, now)
	monKey := monthKey(userID, now)

	dayTok, err := s.store.Get(ctx, tokKey)
	if err != nil {
		return ErrUnavailable
	}
	monTok, err := s.store.Get(ctx, monKey)
	if err != nil {
		return ErrUnavailable
	}
	if (p.dailyTok > 0 && dayTok >= p.dailyTok) || (p.monthTok > 0 && monTok >= p.monthTok) {
		s.recordBlocked(userID, source, "")
		return ErrExceeded
	}

	if p.rpm > 0 {
		n, err := s.store.Incr(ctx, rpmKey(userID, now), 2*time.Minute)
		if err != nil {
			return ErrUnavailable
		}
		if n > int64(p.rpm) {
			return ErrRateLimited
		}
	}

	if p.dailyReq > 0 {
		n, err := s.store.Incr(ctx, reqKey, 48*time.Hour)
		if err != nil {
			return ErrUnavailable
		}
		if n > int64(p.dailyReq) {
			_, _ = s.store.Decr(ctx, reqKey)
			s.recordBlocked(userID, source, "")
			return ErrExceeded
		}
	}
	return nil
}

// RefundRequest 上游失败时退回今日次数。
func (s *Service) RefundRequest(ctx context.Context, userID string) {
	if userID == "" || s.store == nil {
		return
	}
	role, err := s.loadUser(userID)
	if err != nil {
		return
	}
	if s.resolvePolicy(userID, role).unlimited {
		return
	}
	reqKey, _ := dayKey(userID, nowSH())
	_, _ = s.store.Decr(ctx, reqKey)
}

type CommitInput struct {
	UserID           string
	Source           string
	Model            string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	Estimated        bool
	Status           string
	LatencyMS        int64
	TraceID          string
}

func (s *Service) Commit(ctx context.Context, in CommitInput) {
	if in.TotalTokens <= 0 {
		in.TotalTokens = in.PromptTokens + in.CompletionTokens
	}
	if in.Status == "" {
		in.Status = statusOK
	}
	if in.UserID != "" && s.store != nil && in.TotalTokens > 0 && in.Status == statusOK {
		now := nowSH()
		_, tokKey := dayKey(in.UserID, now)
		_, _ = s.store.IncrBy(ctx, tokKey, int64(in.TotalTokens), 48*time.Hour)
		_, _ = s.store.IncrBy(ctx, monthKey(in.UserID, now), int64(in.TotalTokens), 40*24*time.Hour)
	}
	s.writeEvent(in)
}

func (s *Service) recordBlocked(userID, source, model string) {
	s.writeEvent(CommitInput{
		UserID: userID,
		Source: source,
		Model:  model,
		Status: statusBlocked,
	})
}

func (s *Service) writeEvent(in CommitInput) {
	if s.db == nil {
		return
	}
	ev := model.LLMUsageEvent{
		ID:               newID(),
		UserID:           in.UserID,
		Source:           in.Source,
		Model:            in.Model,
		PromptTokens:     in.PromptTokens,
		CompletionTokens: in.CompletionTokens,
		TotalTokens:      in.TotalTokens,
		TokensEstimated:  in.Estimated,
		Status:           in.Status,
		LatencyMS:        in.LatencyMS,
		TraceID:          in.TraceID,
	}
	if err := s.db.Create(&ev).Error; err != nil {
		slog.Warn("llm_usage_events insert failed", "error", err)
	}
}

func (s *Service) Me(ctx context.Context, userID string) (*Snapshot, error) {
	role, err := s.loadUser(userID)
	if err != nil {
		return nil, err
	}
	p := s.resolvePolicy(userID, role)
	now := nowSH()
	snap := &Snapshot{
		Unlimited:      p.unlimited,
		AdminUnlimited: p.unlimited,
		Day: Window{
			RequestsLimit: p.dailyReq,
			TokensLimit:   p.dailyTok,
			ResetsAt:      nextMidnight(now),
		},
		Month: Window{
			TokensLimit: p.monthTok,
			ResetsAt:    nextMonth(now),
		},
	}
	if s.store == nil {
		return snap, ErrUnavailable
	}
	reqKey, tokKey := dayKey(userID, now)
	snap.Day.RequestsUsed, _ = s.store.Get(ctx, reqKey)
	snap.Day.TokensUsed, _ = s.store.Get(ctx, tokKey)
	snap.Month.TokensUsed, _ = s.store.Get(ctx, monthKey(userID, now))
	snap.Warn = warnAt(snap.Day.RequestsUsed, int64(p.dailyReq)) ||
		warnAt(snap.Day.TokensUsed, p.dailyTok) ||
		warnAt(snap.Month.TokensUsed, p.monthTok)
	return snap, nil
}

func warnAt(used, limit int64) bool {
	if limit <= 0 {
		return false
	}
	return used*100 >= limit*80
}

type UsageQuery struct {
	Range   string
	Scope   string
	UserID  string
	IsAdmin bool
}

type UsagePoint struct {
	Date     string `json:"date"`
	Requests int64  `json:"requests"`
	Tokens   int64  `json:"tokens"`
}

type ModelShare struct {
	Model  string `json:"model"`
	Tokens int64  `json:"tokens"`
}

type TopUser struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	Tokens int64  `json:"tokens"`
}

type UsageReport struct {
	Points   []UsagePoint `json:"points"`
	ByModel  []ModelShare `json:"by_model"`
	TopUsers []TopUser    `json:"top_users,omitempty"`
}

func (s *Service) Usage(q UsageQuery) (*UsageReport, error) {
	days := 7
	if q.Range == "30d" {
		days = 30
	}
	start := nowSH().AddDate(0, 0, -days+1)
	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, shanghai)

	db := s.db.Model(&model.LLMUsageEvent{}).Where("created_at >= ?", start)
	targetUser := q.UserID
	if q.IsAdmin && q.Scope == "all" {
		targetUser = ""
	}
	if q.IsAdmin && q.Scope != "all" && q.UserID != "" {
		targetUser = q.UserID
	}
	if !q.IsAdmin {
		targetUser = q.UserID
	}
	if targetUser != "" {
		db = db.Where("user_id = ?", targetUser)
	}

	var events []model.LLMUsageEvent
	if err := db.Find(&events).Error; err != nil {
		return nil, err
	}

	pointsMap := map[string]*UsagePoint{}
	for i := 0; i < days; i++ {
		d := start.AddDate(0, 0, i).Format("2006-01-02")
		pointsMap[d] = &UsagePoint{Date: d}
	}
	modelTok := map[string]int64{}
	userTok := map[string]int64{}

	for _, ev := range events {
		d := ev.CreatedAt.In(shanghai).Format("2006-01-02")
		p := pointsMap[d]
		if p == nil {
			p = &UsagePoint{Date: d}
			pointsMap[d] = p
		}
		if ev.Status == statusOK || ev.Status == "" {
			p.Requests++
			p.Tokens += int64(ev.TotalTokens)
			if ev.Model != "" {
				modelTok[ev.Model] += int64(ev.TotalTokens)
			}
			if ev.UserID != "" {
				userTok[ev.UserID] += int64(ev.TotalTokens)
			}
		}
	}

	out := &UsageReport{Points: make([]UsagePoint, 0, days), ByModel: []ModelShare{}}
	for i := 0; i < days; i++ {
		d := start.AddDate(0, 0, i).Format("2006-01-02")
		out.Points = append(out.Points, *pointsMap[d])
	}
	for m, t := range modelTok {
		out.ByModel = append(out.ByModel, ModelShare{Model: m, Tokens: t})
	}
	if q.IsAdmin && (q.Scope == "all" || targetUser == "") {
		out.TopUsers = s.topUsers(userTok, 10)
	}
	return out, nil
}

func (s *Service) topUsers(tok map[string]int64, n int) []TopUser {
	type pair struct {
		id string
		t  int64
	}
	arr := make([]pair, 0, len(tok))
	for id, t := range tok {
		arr = append(arr, pair{id, t})
	}
	for i := 0; i < len(arr); i++ {
		for j := i + 1; j < len(arr); j++ {
			if arr[j].t > arr[i].t {
				arr[i], arr[j] = arr[j], arr[i]
			}
		}
	}
	if len(arr) > n {
		arr = arr[:n]
	}
	names := map[string]string{}
	ids := make([]string, 0, len(arr))
	for _, p := range arr {
		ids = append(ids, p.id)
	}
	if s.db != nil && len(ids) > 0 {
		var users []model.User
		s.db.Select("id", "name").Where("id IN ?", ids).Find(&users)
		for _, u := range users {
			names[u.ID] = u.Name
		}
	}
	out := make([]TopUser, 0, len(arr))
	for _, p := range arr {
		out = append(out, TopUser{UserID: p.id, Name: names[p.id], Tokens: p.t})
	}
	return out
}

func (s *Service) GetDefaultPolicy() (*model.QuotaPolicy, error) {
	var p model.QuotaPolicy
	err := s.db.Where("user_id IS NULL").First(&p).Error
	if err == gorm.ErrRecordNotFound {
		s.EnsureDefaultPolicy()
		err = s.db.Where("user_id IS NULL").First(&p).Error
	}
	return &p, err
}

func (s *Service) UpdateDefaultPolicy(in model.QuotaPolicy) (*model.QuotaPolicy, error) {
	p, err := s.GetDefaultPolicy()
	if err != nil {
		return nil, err
	}
	p.DailyRequests = in.DailyRequests
	p.DailyTokens = in.DailyTokens
	p.MonthlyTokens = in.MonthlyTokens
	p.RPM = in.RPM
	p.AdminUnlimited = in.AdminUnlimited
	p.Enabled = in.Enabled
	if err := s.db.Save(p).Error; err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) UpsertUserPolicy(userID string, in *model.QuotaPolicy) (*model.QuotaPolicy, error) {
	if in == nil {
		if err := s.db.Where("user_id = ?", userID).Delete(&model.QuotaPolicy{}).Error; err != nil {
			return nil, err
		}
		return nil, nil
	}
	var p model.QuotaPolicy
	err := s.db.Where("user_id = ?", userID).First(&p).Error
	uid := userID
	if err == gorm.ErrRecordNotFound {
		p = model.QuotaPolicy{ID: newID(), UserID: &uid}
	} else if err != nil {
		return nil, err
	}
	p.DailyRequests = in.DailyRequests
	p.DailyTokens = in.DailyTokens
	p.MonthlyTokens = in.MonthlyTokens
	p.RPM = in.RPM
	p.AdminUnlimited = in.AdminUnlimited
	p.Enabled = in.Enabled
	if p.UserID == nil {
		p.UserID = &uid
	}
	if err := s.db.Save(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func WriteError(c *gin.Context, err error) {
	switch {
	case err == nil:
		return
	case err == ErrRateLimited:
		response.FailWithHTTPStatus(c, 429, response.CodeRateLimitExceeded, "")
	case err == ErrExceeded:
		response.FailWithHTTPStatus(c, 429, response.CodeUserQuotaExceeded, "")
	case err == ErrUnavailable:
		response.FailWithHTTPStatus(c, 503, response.CodeCacheError, "额度服务暂不可用，请稍后重试")
	default:
		response.FailWithHTTPStatus(c, 500, response.CodeSystemError, err.Error())
	}
}

func TraceID(c *gin.Context) string {
	return trace.GetTraceIDOrGenerate(c)
}
