package routes

import (
	"pathshala/controllers"
	"pathshala/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupCollegeTypeRoutes(r *gin.Engine, db *gorm.DB) {
	collegeTypeGroup := r.Group("/api/college_types")

	collegeTypeGroup.Use(middlewares.AuthMiddleware(), middlewares.RoleMiddleware("admin", "teacher"))

	// Create a new college type
	collegeTypeGroup.POST("/", func(c *gin.Context) {
		controllers.CreateCollegeType(c, db)
	})

	// Get all college types
	collegeTypeGroup.GET("/", func(c *gin.Context) {
		controllers.GetAllCollegeTypes(c, db)
	})

	// Get a single college type by ID
	collegeTypeGroup.GET("/:id", func(c *gin.Context) {
		controllers.GetCollegeTypeByID(c, db)
	})

	// Update a college type by ID
	collegeTypeGroup.PUT("/:id", func(c *gin.Context) {
		controllers.UpdateCollegeType(c, db)
	})

	// Delete a college type by ID
	collegeTypeGroup.DELETE("/:id", func(c *gin.Context) {
		controllers.DeleteCollegeType(c, db)
	})

}
