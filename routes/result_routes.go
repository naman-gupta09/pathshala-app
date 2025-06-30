package routes

import (
	"pathshala/controllers"
	"pathshala/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupResultRoutes(router *gin.Engine) {
	results := router.Group("api/results").Use(middlewares.AuthMiddleware(), middlewares.RoleMiddleware("admin", "teacher"))
	results.POST("/", controllers.SubmitResults)
	results.GET("/", controllers.GetResults)
}
