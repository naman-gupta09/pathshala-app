package routes

import (
	"pathshala/controllers"
	"pathshala/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupMacroCategoryRoutes(router *gin.Engine) {
	macro := router.Group("/api/macro-categories").Use(middlewares.AuthMiddleware(), middlewares.RoleMiddleware("admin", "teacher"))
	{
		macro.POST("/", controllers.CreateMacroCategory)
		macro.GET("/", controllers.GetAllMacroCategories)
		macro.PUT("/:id", controllers.UpdateMacroCategory)
		macro.DELETE("/:id", controllers.DeleteMacroCategory)
	}
}
