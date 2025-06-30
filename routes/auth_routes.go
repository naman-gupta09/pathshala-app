package routes

import (
	"pathshala/controllers"
	middlewares "pathshala/middlewares"
	"time"

	"github.com/gin-gonic/gin"
)

func SetupAuthRoutes(r *gin.Engine) {
	r.POST("/register", middlewares.AuthMiddleware(), middlewares.RoleMiddleware("admin", "teacher"), middlewares.TimeoutMiddleware(5*time.Second), controllers.Register)
	// r.POST("/register", controllers.Register)
	r.POST("/login", middlewares.TimeoutMiddleware(5*time.Second), controllers.Login)
	r.POST("/refresh", controllers.RefreshToken)
	r.POST("/forgot-password", middlewares.TimeoutMiddleware(5*time.Second), controllers.ForgotPassword)
	r.POST("/reset-password", controllers.ResetPassword)
	r.POST("/logout", controllers.Logout)
}
