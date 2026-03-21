package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/orjrs/gateway/pkg/orjrs/gw/middleware"
	"golang.org/x/crypto/bcrypt"
)

// User represents a user model
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// In-memory user store (replace with database in production)
var users = make(map[string]*User)

func init() {
	// 创建默认管理员账号
	adminPassword := "admin123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		panic("Failed to hash admin password: " + err.Error())
	}
	admin := &User{
		ID:        generateID(),
		Email:     "admin@cogniforge.local",
		Name:      "admin",
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	users[admin.Email] = admin
}

// RegisterRequest represents registration request
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
}

// LoginRequest represents login request
type LoginRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// Register handles user registration
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already exists
	if _, exists := users[req.Email]; exists {
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
	user := &User{
		ID:        generateID(),
		Email:     req.Email,
		Name:      req.Name,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	users[req.Email] = user

	// Generate token
	token, err := generateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, AuthResponse{
		Token: token,
		User:  *user,
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

	// Find user by email or username
	var user *User
	var exists bool
	if req.Email != "" {
		user, exists = users[req.Email]
	} else if req.Username != "" {
		for _, u := range users {
			if u.Name == req.Username {
				user = u
				exists = true
				break
			}
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请输入邮箱或用户名"})
		return
	}

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate token
	token, err := generateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Token: token,
		User:  *user,
	})
}

// Logout handles user logout
func Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// GetCurrentUser returns current user info
func GetCurrentUser(c *gin.Context) {
	userID := c.GetString("user_id")

	// Find user by ID
	var currentUser *User
	for _, user := range users {
		if user.ID == userID {
			currentUser = user
			break
		}
	}

	if currentUser == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, currentUser)
}

// Placeholder handlers for other endpoints
func ListUsers(c *gin.Context)       { c.JSON(http.StatusOK, users) }
func GetUser(c *gin.Context)         { c.JSON(http.StatusOK, gin.H{"message": "Get user"}) }
func UpdateUser(c *gin.Context)      { c.JSON(http.StatusOK, gin.H{"message": "Update user"}) }
func DeleteUser(c *gin.Context)      { c.JSON(http.StatusOK, gin.H{"message": "Delete user"}) }
func ListApiKeys(c *gin.Context)     { c.JSON(http.StatusOK, gin.H{"keys": []interface{}{}}) }
func CreateApiKey(c *gin.Context)    { c.JSON(http.StatusOK, gin.H{"key": "sk-" + generateID()}) }
func DeleteApiKey(c *gin.Context)    { c.JSON(http.StatusOK, gin.H{"message": "API key deleted"}) }
func ListModels(c *gin.Context)      { c.JSON(http.StatusOK, gin.H{"models": []interface{}{}}) }
func GetModel(c *gin.Context)        { c.JSON(http.StatusOK, gin.H{"message": "Get model"}) }
func Chat(c *gin.Context)            { c.JSON(http.StatusOK, gin.H{"message": "Chat"}) }
func ChatStream(c *gin.Context)      { c.JSON(http.StatusOK, gin.H{"message": "Chat stream"}) }
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
func generateToken(user *User) (string, error) {
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
