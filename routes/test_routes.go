package routes

import (
	"pathshala/controllers"
	"pathshala/middlewares"

	"github.com/gin-gonic/gin"
)

// func RegisterTestRoutes(r *gin.Engine) {
// 	testGroup := r.Group("/test")
// 	{
// 		testGroup.POST("/", controllers.CreateTest)
// 	}
// }

func SetupTestRoutes(router *gin.Engine) {

	testRoutes := router.Group("/aditi/tests").Use(middlewares.AuthMiddleware(), middlewares.RoleMiddleware("admin", "teacher"))
	testRoutes.GET("/", controllers.GetTests)
	testRoutes.POST("/", controllers.CreateTest)
	testRoutes.POST("/send-test", controllers.SendTest)
	testRoutes.DELETE("/:id", controllers.DeleteTest)
	testRoutes.GET("/states", controllers.GetStates)
	testRoutes.GET("/colleges", controllers.GetCollegesByState)

	// protected := router.Group("/").Use(middlewares.AuthMiddleware(), middlewares.RoleMiddleware("admin", "teacher"))
	// // protected.Use(middlewares.MockAuthMiddleware()) // <-- Use mock here
	// {
	// 	protected.POST("/tests", controllers.CreateTest)
	// 	protected.POST("/send-test", controllers.SendTest)
	// 	protected.DELETE("/tests/:id", controllers.DeleteTest)
	// }
}
