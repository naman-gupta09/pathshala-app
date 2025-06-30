package routes

import (
	"pathshala/controllers"
	"pathshala/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupProfileRoutes(r *gin.Engine) {

	profile := r.Group("/api/profile").Use(middlewares.AuthMiddleware(), middlewares.RoleMiddleware("admin", "teacher"))

	profile.GET("/", controllers.GetProfile)
	profile.PUT("/", controllers.UpdateProfile)

}
