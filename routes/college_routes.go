package routes

import (
	"pathshala/controllers"
	middlewares "pathshala/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupCollegeRoutes(r *gin.Engine, db *gorm.DB) {
	collegeGroup := r.Group("/api/colleges")

	collegeGroup.Use(middlewares.AuthMiddleware(), middlewares.RoleMiddleware("admin", "teacher"))

	collegeGroup.POST("/", func(c *gin.Context) {
		controllers.CreateCollege(c, db)
	})
	collegeGroup.GET("/", func(c *gin.Context) {
		controllers.GetColleges(c, db)
	})
	collegeGroup.PUT("/:id", func(c *gin.Context) {
		controllers.UpdateCollege(c, db)
	})
	collegeGroup.DELETE("/:id", func(c *gin.Context) {
		controllers.DeleteCollege(c, db)
	})
}
