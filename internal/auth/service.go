package auth

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"cogniforge/internal/database"
	"cogniforge/internal/mail"
	"cogniforge/internal/middleware"
	"cogniforge/internal/model"
	"cogniforge/internal/util"
)

type AuthService struct {
	db     *gorm.DB
	rdb    *redis.Client
	mailer mail.Sender
}

func NewAuthService() *AuthService {
	return &AuthService{db: database.DB, mailer: mail.Nop{}}
}

func NewAuthServiceWithDeps(db *gorm.DB, rdb *redis.Client, mailer mail.Sender) *AuthService {
	if mailer == nil {
		mailer = mail.Nop{}
	}
	return &AuthService{db: db, rdb: rdb, mailer: mailer}
}

// PurgeSeedLocalAdmin 物理删除种子账号 admin@cogniforge.local（含软删行）。
// 必须在 AutoMigrate 加 uk_users_name 之前调用，避免与真实 admin 用户名冲突。
func PurgeSeedLocalAdmin(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	res := db.Unscoped().Where("email = ?", "admin@cogniforge.local").Delete(&model.User{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected > 0 {
		slog.Info("purged seed admin account", "email", "admin@cogniforge.local", "rows", res.RowsAffected)
	}
	return nil
}

// PrepareUserNameUniqueness AutoMigrate 加 uk_users_name 前清理冲突：
// 1) 物理删种子邮箱；2) 软删行若仍占用 name，改成释放名；3) 活跃同名保留最早一条，其余改名后软删。
func PrepareUserNameUniqueness(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	if err := PurgeSeedLocalAdmin(db); err != nil {
		return err
	}

	// 软删行仍占 name 时释放（旧版软删未改名）
	var softDeleted []model.User
	if err := db.Unscoped().Where("deleted_at IS NOT NULL").Find(&softDeleted).Error; err != nil {
		return err
	}
	for _, u := range softDeleted {
		if strings.Contains(u.Name, "__del_") {
			continue
		}
		stamp := time.Now().Unix()
		short := u.ID
		if len(short) > 8 {
			short = short[:8]
		}
		freed := fmt.Sprintf("%s__del_%s_%d", u.Name, short, stamp)
		if len(freed) > 100 {
			freed = freed[:100]
		}
		if err := db.Unscoped().Model(&u).Update("name", freed).Error; err != nil {
			return err
		}
		slog.Info("freed soft-deleted username", "id", u.ID, "name", freed)
	}

	var active []model.User
	if err := db.Where("deleted_at IS NULL").Order("created_at ASC").Find(&active).Error; err != nil {
		return err
	}
	seen := map[string]string{} // name -> kept id
	for _, u := range active {
		name := strings.TrimSpace(u.Name)
		if name == "" {
			continue
		}
		if keptID, ok := seen[name]; ok && keptID != u.ID {
			stamp := time.Now().Unix()
			short := u.ID
			if len(short) > 8 {
				short = short[:8]
			}
			freedName := fmt.Sprintf("%s__dup_%s_%d", name, short, stamp)
			if len(freedName) > 100 {
				freedName = freedName[:100]
			}
			freedEmail := fmt.Sprintf("dup+%s.%d@invalid.local", short, stamp)
			if err := db.Model(&u).Updates(map[string]interface{}{
				"name":  freedName,
				"email": freedEmail,
			}).Error; err != nil {
				return err
			}
			if err := db.Delete(&u).Error; err != nil {
				return err
			}
			slog.Warn("soft-deleted duplicate username", "kept_id", keptID, "removed_id", u.ID, "name", name)
			continue
		}
		seen[name] = u.ID
	}
	return nil
}

// InitDefaultAdmin 仅在「库里一个用户都没有」时创建首个管理员。
// 不再每次启动补种 admin@cogniforge.local（会与用户自建同名 admin 冲突）。
func (s *AuthService) InitDefaultAdmin() error {
	var count int64
	if err := s.db.Model(&model.User{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash admin password: %w", err)
	}
	admin := model.User{
		ID:        generateID(),
		Email:     "admin@cogniforge.local",
		Name:      "admin",
		Password:  string(hashedPassword),
		Role:      "admin",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.db.Create(&admin).Error; err != nil {
		return fmt.Errorf("failed to create default admin: %w", err)
	}

	settings := model.UserSettings{
		ID:        generateID(),
		UserID:    admin.ID,
		AvatarURL: "",
		Theme:     "light",
		Language:  "zh-CN",
		Timezone:  "Asia/Shanghai",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	s.db.Create(&settings)
	slog.Info("created first-run default admin", "email", admin.Email, "name", admin.Name)
	return nil
}

// Register 注册新用户
func (s *AuthService) Register(c *gin.Context) (*AuthData, error) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	if !isValidEmail(req.Email) {
		return nil, fmt.Errorf("请输入有效的邮箱地址")
	}
	if len(req.Password) < 6 {
		return nil, fmt.Errorf("密码至少6位")
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))
	name := util.NormalizeUsername(req.Name)
	if err := util.ValidateUsername(name); err != nil {
		return nil, err
	}

	var existing model.User
	if err := s.db.Where("email = ?", email).First(&existing).Error; err == nil {
		return nil, fmt.Errorf("该邮箱已被注册")
	}
	if err := s.db.Where("name = ?", name).First(&existing).Error; err == nil {
		return nil, fmt.Errorf("该用户名已被使用")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败")
	}

	user := model.User{
		ID:        generateID(),
		Email:     email,
		Name:      name,
		Password:  string(hashedPassword),
		Role:      "user",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.db.Create(&user).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, fmt.Errorf("邮箱或用户名已被使用")
		}
		return nil, fmt.Errorf("创建用户失败")
	}

	settings := model.UserSettings{
		ID:        generateID(),
		UserID:    user.ID,
		AvatarURL: "",
		Theme:     "light",
		Language:  "zh-CN",
		Timezone:  "Asia/Shanghai",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	s.db.Create(&settings)

	token, err := s.generateToken(&user)
	if err != nil {
		return nil, fmt.Errorf("生成 Token 失败")
	}

	return &AuthData{Token: token, User: ToUserResponse(&user)}, nil
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique") || strings.Contains(msg, "duplicate")
}

// Login 用户登录
func (s *AuthService) Login(c *gin.Context) (*AuthData, error) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	if req.Password == "" {
		return nil, fmt.Errorf("密码不能为空")
	}

	var user model.User
	var err error

	if req.Email != "" {
		err = s.db.Where("email = ?", req.Email).First(&user).Error
	} else if req.Username != "" {
		err = s.db.Where("name = ?", req.Username).First(&user).Error
	} else {
		return nil, fmt.Errorf("请输入邮箱或用户名")
	}

	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("用户名或密码错误")
	}
	if err != nil {
		return nil, fmt.Errorf("数据库查询失败")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("用户名或密码错误")
	}

	token, err := s.generateToken(&user)
	if err != nil {
		return nil, fmt.Errorf("生成 Token 失败")
	}

	session := model.UserSession{
		ID:        generateID(),
		UserID:    user.ID,
		TokenID:   user.ID,
		UserAgent: c.GetHeader("User-Agent"),
		IPAddress: c.ClientIP(),
		Device:    parseDeviceFromUA(c.GetHeader("User-Agent")),
		ExpiresAt: time.Now().Add(24 * time.Hour),
		LastUsed:  time.Now(),
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	slog.Debug("CreateSession",
		"userID", user.ID,
		"device", session.Device,
		"ip", session.IPAddress,
	)

	s.db.Create(&session)

	return &AuthData{Token: token, User: ToUserResponse(&user)}, nil
}

// Logout 用户登出
func (s *AuthService) Logout(c *gin.Context) error {
	return nil
}

// GetCurrentUser 获取当前用户
func (s *AuthService) GetCurrentUser(userID string) (*model.User, error) {
	var user model.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败")
	}
	return &user, nil
}

// ListApiKeys 获取用户的 API Key 列表
func (s *AuthService) ListApiKeys(userID string) ([]model.ApiKey, error) {
	var keys []model.ApiKey
	if err := s.db.Where("user_id = ?", userID).Find(&keys).Error; err != nil {
		return nil, fmt.Errorf("查询 API Key 失败")
	}
	return keys, nil
}

// CreateApiKey 创建新的 API Key
func (s *AuthService) CreateApiKey(userID string, req *ApiKeyRequest) (*model.ApiKey, error) {
	key := "sk-" + generateID()
	apiKey := model.ApiKey{
		ID:        generateID(),
		UserID:    userID,
		Name:      req.Name,
		Key:       key,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.db.Create(&apiKey).Error; err != nil {
		return nil, fmt.Errorf("创建 API Key 失败")
	}
	return &apiKey, nil
}

// DeleteApiKey 删除 API Key
func (s *AuthService) DeleteApiKey(userID, keyID string) error {
	var apiKey model.ApiKey
	if err := s.db.Where("id = ?", keyID).First(&apiKey).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("API Key 不存在")
		}
		return fmt.Errorf("数据库查询失败")
	}

	if apiKey.UserID != userID {
		return fmt.Errorf("无权删除此 API Key")
	}

	if err := s.db.Delete(&apiKey).Error; err != nil {
		return fmt.Errorf("删除 API Key 失败")
	}
	return nil
}

// generateToken 生成 JWT Token
func (s *AuthService) generateToken(user *model.User) (string, error) {
	claims := &middleware.Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(middleware.JWTSecret)
}

// ============ 辅助函数 ============

func isValidEmail(email string) bool {
	atIndex := -1
	dotAfterAt := false
	for i, ch := range email {
		if ch == '@' {
			if atIndex != -1 {
				return false
			}
			atIndex = i
		} else if ch == '.' && atIndex != -1 {
			dotAfterAt = true
		}
	}
	return atIndex > 0 && dotAfterAt && atIndex < len(email)-1
}

// parseDeviceFromUA 从 User-Agent 解析设备类型和浏览器
func parseDeviceFromUA(ua string) string {
	if ua == "" {
		return "Unknown Device"
	}
	uaLower := strings.ToLower(ua)

	browser := ""
	if strings.Contains(uaLower, "edg/") || strings.Contains(uaLower, "edge/") {
		browser = "Edge"
	} else if strings.Contains(uaLower, "chrome/") && !strings.Contains(uaLower, "chromium/") {
		browser = "Chrome"
	} else if strings.Contains(uaLower, "firefox/") {
		browser = "Firefox"
	} else if strings.Contains(uaLower, "safari/") && !strings.Contains(uaLower, "chrome/") {
		browser = "Safari"
	} else if strings.Contains(uaLower, "opera/") || strings.Contains(uaLower, "opr/") {
		browser = "Opera"
	}

	deviceType := "Desktop"
	if strings.Contains(uaLower, "mobile") || strings.Contains(uaLower, "iphone") || strings.Contains(uaLower, "ipod") {
		deviceType = "Mobile"
	} else if strings.Contains(uaLower, "android") && !strings.Contains(uaLower, "mobile") {
		deviceType = "Android Tablet"
	} else if strings.Contains(uaLower, "tablet") || strings.Contains(uaLower, "ipad") {
		deviceType = "iPad"
	}

	if browser != "" {
		return fmt.Sprintf("%s on %s", browser, deviceType)
	}
	return deviceType
}

// generateID 生成唯一ID
func generateID() string {
	return uuid.New().String()
}
