package routes

import (
	"pathshala/controllers"
	"pathshala/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupCategoryRoutes(router *gin.Engine) {
	category := router.Group("/api/categories").Use(middlewares.AuthMiddleware(), middlewares.RoleMiddleware("admin", "teacher"))
	{
		category.POST("/", controllers.AddCategory)
		category.GET("/", controllers.GetAllCategories)
		category.GET("/:id", controllers.GetCategoryByID)
		category.PUT("/:id", controllers.UpdateCategory)
		category.DELETE("/:id", controllers.DeleteCategory)
	}
}
