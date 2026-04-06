package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"cogniforge/internal/database"
	"cogniforge/internal/middleware"
	"cogniforge/internal/model"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type ApiKeyRequest struct {
	Name string `json:"name"`
}

type AuthData struct {
	Token string     `json:"token"`
	User  model.User `json:"user"`
}

func InitDefaultAdmin() {
	var admin model.User
	err := database.DB.Where("email = ?", "admin@cogniforge.local").First(&admin).Error
	if err == gorm.ErrRecordNotFound {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if err != nil {
			panic("Failed to hash admin password: " + err.Error())
		}
		admin = model.User{
			ID:        generateID(),
			Email:     "admin@cogniforge.local",
			Name:      "admin",
			Password:  string(hashedPassword),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := database.DB.Create(&admin).Error; err != nil {
			panic("Failed to create default admin: " + err.Error())
		}
	}
}

func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		model.BadRequest(c, err.Error())
		return
	}

	if req.Email == "" || !isValidEmail(req.Email) {
		model.BadRequest(c, "请输入有效的邮箱地址")
		return
	}
	if len(req.Password) < 6 {
		model.BadRequest(c, "密码至少6位")
		return
	}
	if req.Name == "" {
		model.BadRequest(c, "请输入用户名")
		return
	}

	var existing model.User
	if err := database.DB.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		model.Fail(c, http.StatusConflict, "该邮箱已被注册")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		model.InternalError(c, "密码加密失败")
		return
	}

	user := model.User{
		ID:        generateID(),
		Email:     req.Email,
		Name:      req.Name,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := database.DB.Create(&user).Error; err != nil {
		model.InternalError(c, "创建用户失败")
		return
	}

	token, err := generateToken(&user)
	if err != nil {
		model.InternalError(c, "生成 Token 失败")
		return
	}

	model.Created(c, AuthData{Token: token, User: user})
}

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		model.BadRequest(c, err.Error())
		return
	}

	if req.Password == "" {
		model.BadRequest(c, "密码不能为空")
		return
	}

	var user model.User
	var err error

	if req.Email != "" {
		err = database.DB.Where("email = ?", req.Email).First(&user).Error
	} else if req.Username != "" {
		err = database.DB.Where("name = ?", req.Username).First(&user).Error
	} else {
		model.BadRequest(c, "请输入邮箱或用户名")
		return
	}

	if err == gorm.ErrRecordNotFound {
		model.Unauthorized(c, "用户名或密码错误")
		return
	}
	if err != nil {
		model.InternalError(c, "数据库查询失败")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		model.Unauthorized(c, "用户名或密码错误")
		return
	}

	token, err := generateToken(&user)
	if err != nil {
		model.InternalError(c, "生成 Token 失败")
		return
	}

	model.Success(c, AuthData{Token: token, User: user})
}

func Logout(c *gin.Context) {
	model.SuccessWithMessage(c, nil, "已退出登录")
}

func GetCurrentUser(c *gin.Context) {
	userID := c.GetString("user_id")
	var user model.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err == gorm.ErrRecordNotFound {
		model.NotFound(c, "用户不存在")
		return
	}
	model.Success(c, user)
}

func ListUsers(c *gin.Context) {
	var users []model.User
	database.DB.Find(&users)
	model.Success(c, users)
}

func GetUser(c *gin.Context)    { model.Success(c, gin.H{"message": "Get user"}) }
func UpdateUser(c *gin.Context) { model.Success(c, gin.H{"message": "Update user"}) }
func DeleteUser(c *gin.Context) { model.Success(c, gin.H{"message": "Delete user"}) }

func ListApiKeys(c *gin.Context) {
	userID := c.GetString("user_id")
	var keys []model.ApiKey
	database.DB.Where("user_id = ?", userID).Find(&keys)
	model.Success(c, gin.H{"keys": keys})
}

func CreateApiKey(c *gin.Context) {
	userID := c.GetString("user_id")
	var req ApiKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		model.BadRequest(c, err.Error())
		return
	}
	key := "sk-" + generateID()
	apiKey := model.ApiKey{
		ID:        generateID(),
		UserID:    userID,
		Name:      req.Name,
		Key:       key,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := database.DB.Create(&apiKey).Error; err != nil {
		model.InternalError(c, "创建 API Key 失败")
		return
	}
	model.Created(c, gin.H{"id": apiKey.ID, "name": apiKey.Name, "key": apiKey.Key, "created_at": apiKey.CreatedAt})
}

func DeleteApiKey(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	var apiKey model.ApiKey
	if err := database.DB.Where("id = ?", id).First(&apiKey).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			model.NotFound(c, "API Key 不存在")
		} else {
			model.InternalError(c, "数据库查询失败")
		}
		return
	}

	if apiKey.UserID != userID {
		model.Forbidden(c, "无权删除此 API Key")
		return
	}

	if err := database.DB.Delete(&apiKey).Error; err != nil {
		model.InternalError(c, "删除 API Key 失败")
		return
	}
	model.SuccessWithMessage(c, nil, "API Key 已撤销")
}

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

func generateToken(user *model.User) (string, error) {
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

func generateID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
		time.Sleep(time.Nanosecond)
	}
	return string(b)
}
