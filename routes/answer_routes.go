package routes

import (
	"pathshala/controllers"
	"pathshala/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupAnswerRoutes(router *gin.Engine) {
	answers := router.Group("/api/answers").Use(middlewares.AuthMiddleware(), middlewares.RoleMiddleware("admin", "teacher"))
	{
		answers.POST("/", controllers.SubmitAnswers)
	}
}
