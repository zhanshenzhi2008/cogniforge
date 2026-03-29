package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/orjrs/gateway/pkg/orjrs/gw/database"
	"github.com/orjrs/gateway/pkg/orjrs/gw/middleware"
	"github.com/orjrs/gateway/pkg/orjrs/gw/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// RegisterRequest represents registration request
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// LoginRequest represents login request
type LoginRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	Token string     `json:"token"`
	User  model.User `json:"user"`
}

// ApiKeyRequest represents API key creation request
type ApiKeyRequest struct {
	Name string `json:"name" binding:"required"`
}

// InitDefaultAdmin creates the default admin user if not exists
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

// Register handles user registration
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate email format
	if req.Email == "" || !isValidEmail(req.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请输入有效的邮箱地址"})
		return
	}

	// Validate password length
	if len(req.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "密码至少6位"})
		return
	}

	// Validate name
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请输入用户名"})
		return
	}

	// Check if user already exists
	var existing model.User
	if err := database.DB.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create user
	user := model.User{
		ID:        generateID(),
		Email:     req.Email,
		Name:      req.Name,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Generate token
	token, err := generateToken(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, AuthResponse{
		Token: token,
		User:  user,
	})
}

// Login handles user login
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "密码不能为空"})
		return
	}

	var user model.User
	var err error

	if req.Email != "" {
		err = database.DB.Where("email = ?", req.Email).First(&user).Error
	} else if req.Username != "" {
		err = database.DB.Where("name = ?", req.Username).First(&user).Error
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请输入邮箱或用户名"})
		return
	}

	if err == gorm.ErrRecordNotFound {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate token
	token, err := generateToken(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Token: token,
		User:  user,
	})
}

// Logout handles user logout
func Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// GetCurrentUser returns current user info
func GetCurrentUser(c *gin.Context) {
	userID := c.GetString("user_id")

	var user model.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err == gorm.ErrRecordNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// ListUsers returns all users
func ListUsers(c *gin.Context) {
	var users []model.User
	database.DB.Find(&users)
	c.JSON(http.StatusOK, users)
}

func GetUser(c *gin.Context)    { c.JSON(http.StatusOK, gin.H{"message": "Get user"}) }
func UpdateUser(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "Update user"}) }
func DeleteUser(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "Delete user"}) }

func ListApiKeys(c *gin.Context) {
	userID := c.GetString("user_id")
	var keys []model.ApiKey
	database.DB.Where("user_id = ?", userID).Find(&keys)
	c.JSON(http.StatusOK, gin.H{"keys": keys})
}

func CreateApiKey(c *gin.Context) {
	userID := c.GetString("user_id")
	var req ApiKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create API key"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"key": key, "id": apiKey.ID, "name": apiKey.Name, "created_at": apiKey.CreatedAt})
}

func DeleteApiKey(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	var apiKey model.ApiKey
	if err := database.DB.Where("id = ?", id).First(&apiKey).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "API key not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	// Verify ownership: only the owner can delete their API key
	if apiKey.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to delete this API key"})
		return
	}

	if err := database.DB.Delete(&apiKey).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete API key"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "API key revoked successfully"})
}

func GetModel(c *gin.Context)        { c.JSON(http.StatusOK, gin.H{"message": "Get model"}) }
func ListAgents(c *gin.Context)      { c.JSON(http.StatusOK, gin.H{"agents": []interface{}{}}) }
func CreateAgent(c *gin.Context)     { c.JSON(http.StatusOK, gin.H{"message": "Create agent"}) }
func GetAgent(c *gin.Context)        { c.JSON(http.StatusOK, gin.H{"message": "Get agent"}) }
func UpdateAgent(c *gin.Context)     { c.JSON(http.StatusOK, gin.H{"message": "Update agent"}) }
func DeleteAgent(c *gin.Context)     { c.JSON(http.StatusOK, gin.H{"message": "Delete agent"}) }
func AgentChat(c *gin.Context)       { c.JSON(http.StatusOK, gin.H{"message": "Agent chat"}) }
func ListWorkflows(c *gin.Context)   { c.JSON(http.StatusOK, gin.H{"workflows": []interface{}{}}) }
func CreateWorkflow(c *gin.Context)  { c.JSON(http.StatusOK, gin.H{"message": "Create workflow"}) }
func GetWorkflow(c *gin.Context)     { c.JSON(http.StatusOK, gin.H{"message": "Get workflow"}) }
func UpdateWorkflow(c *gin.Context)  { c.JSON(http.StatusOK, gin.H{"message": "Update workflow"}) }
func DeleteWorkflow(c *gin.Context)  { c.JSON(http.StatusOK, gin.H{"message": "Delete workflow"}) }
func ExecuteWorkflow(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "Execute workflow"}) }
func ListKnowledgeBases(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"knowledge_bases": []interface{}{}})
}
func CreateKnowledgeBase(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Create knowledge base"})
}
func GetKnowledgeBase(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "Get knowledge base"}) }
func UpdateKnowledgeBase(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Update knowledge base"})
}
func DeleteKnowledgeBase(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Delete knowledge base"})
}
func UploadDocument(c *gin.Context)  { c.JSON(http.StatusOK, gin.H{"message": "Upload document"}) }
func ListDocuments(c *gin.Context)   { c.JSON(http.StatusOK, gin.H{"documents": []interface{}{}}) }
func DeleteDocument(c *gin.Context)  { c.JSON(http.StatusOK, gin.H{"message": "Delete document"}) }
func SearchKnowledge(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "Search knowledge"}) }

// Helper functions
func isValidEmail(email string) bool {
	// Basic email format validation
	atIndex := -1
	dotAfterAt := false
	for i, ch := range email {
		if ch == '@' {
			if atIndex != -1 {
				return false // multiple @
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
