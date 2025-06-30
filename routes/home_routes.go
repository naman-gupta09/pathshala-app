package routes

import (
	"pathshala/controllers"

	"pathshala/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupHomeRouter(router *gin.Engine) {
	api := router.Group("/api/home").Use(middlewares.AuthMiddleware(), middlewares.RoleMiddleware("admin", "teacher"))

	api.GET("/stats", controllers.GetHomeStats)

	api.GET("/teacher/stats", controllers.GetTeacherHomeStats)
}
