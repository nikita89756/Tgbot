package server

import (
	"bot/internal/middleware"
	"bot/internal/models"
	"bot/internal/storage"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Server struct {
	host     string
	port     string
	storage  storage.Storage
	jwtToken string
	tokenTTL time.Duration
	logger   *zap.Logger
}

func NewServer(host,port string, storage storage.Storage, jwtToken string, TokenTTL time.Duration, logger *zap.Logger) *Server {
	return &Server{host: host, port: port, storage: storage, jwtToken: jwtToken, tokenTTL: TokenTTL, logger: logger}
}

func (s *Server) Start() error {
	m := middleware.Middleware{s.jwtToken, s.logger}
	engine := gin.New()
	engine.Use(gin.Logger())
	engine.Use(middleware.CORSMiddleware())
	engine.POST("/login", s.loginHandler)
	api := engine.Group("/api", m.AuthMiddleware())
	{
	 	api.GET("/items", s.getItemsHandler)
	 	api.POST("/items", s.postItemsHandler)
	}

	err := engine.Run(fmt.Sprintf("%s:%s", s.host, s.port))
	return err
}

type signinUser struct {
	Email    string `json:"email" example:"user@example.com" binding:"required"`
	Password string `json:"password" example:"password123" binding:"required"`
}

func (s *Server) loginHandler(ctx *gin.Context) {
	var user signinUser
	if err := ctx.BindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	s.logger.Info("Login attempt", zap.String("email", user.Email))
	u, err := s.storage.GetUserByEmail(user.Email, user.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	token, err := middleware.GenerateToken(user.Email, user.Password, u.ID, s.jwtToken, s.tokenTTL)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}
	s.logger.Info("Login successful", zap.String("email", user.Email), zap.String("token", token))

	// Отправляем токен в заголовке Authorization
	ctx.JSON(http.StatusOK, gin.H{
					"message": "Login successful",
 					"token": token,})
}

func (s *Server) postItemsHandler(ctx *gin.Context){
	var item models.Item
	if err := ctx.BindJSON(&item); err != nil {
		s.logger.Error("Failed to bind JSON", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	err := s.storage.InsertItem(item.Name, item.Price, item.Multiplier)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Item added successfully"})
}

func (s *Server) getItemsHandler(ctx *gin.Context){
	items, err := s.storage.GetItems()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, items)
}