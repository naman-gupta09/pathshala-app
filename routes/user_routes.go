package routes

import (
	"pathshala/controllers"
	middlewares "pathshala/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupUserRoutes(r *gin.Engine, db *gorm.DB) {
	userGroup := r.Group("/api/users")

	userGroup.Use(middlewares.AuthMiddleware(), middlewares.RoleMiddleware("admin", "teacher"))
	{
		// Common user management
		userGroup.GET("/students", func(c *gin.Context) {
			controllers.GetUsersByRole(c, db, "student")
		})
		userGroup.GET("/teachers", func(c *gin.Context) {
			controllers.GetUsersByRole(c, db, "teacher")
		})
		userGroup.PUT("/:id", func(c *gin.Context) {
			controllers.UpdateUser(c, db)
		})
		userGroup.DELETE("/:id", func(c *gin.Context) {
			controllers.DeleteUser(c, db)
		})

		// New role-based user creation endpoints
		userGroup.POST("/student", func(c *gin.Context) {
			controllers.CreateStudent(c, db)
		})
		userGroup.POST("/teacher", func(c *gin.Context) {
			controllers.CreateTeacher(c, db)
		})
	}
}
