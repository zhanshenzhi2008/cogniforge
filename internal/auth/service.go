package auth

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"cogniforge/internal/database"
	"cogniforge/internal/middleware"
	"cogniforge/internal/model"
)

type AuthService struct {
	db *gorm.DB
}

func NewAuthService() *AuthService {
	return &AuthService{db: database.DB}
}

// InitDefaultAdmin 初始化默认管理员
func (s *AuthService) InitDefaultAdmin() error {
	var admin model.User
	err := s.db.Where("email = ?", "admin@cogniforge.local").First(&admin).Error
	if err == gorm.ErrRecordNotFound {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash admin password: %w", err)
		}
		admin = model.User{
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
	}
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
	if req.Name == "" {
		return nil, fmt.Errorf("请输入用户名")
	}

	var existing model.User
	if err := s.db.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		return nil, fmt.Errorf("该邮箱已被注册")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败")
	}

	user := model.User{
		ID:        generateID(),
		Email:     req.Email,
		Name:      req.Name,
		Password:  string(hashedPassword),
		Role:      "user",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.db.Create(&user).Error; err != nil {
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
