package routes

// import (
// 	"pathshala/controllers"
// 	"pathshala/middlewares"
// 	"pathshala/services"

// 	"github.com/gin-gonic/gin"
// )

// // routes/survey_routes.go
// func SetupSurveyRoutes(rg *gin.RouterGroup, surveyService *services.SurveyService) {
// 	surveyController := controllers.NewSurveyController(surveyService)

// 	// Create a sub-group for surveys and apply middleware in order.
// 	surveys := rg.Group("/surveys")
// 	// First, inject mock JWT values (userRole and userID).
// 	surveys.Use(middlewares.MockJWT("teacher"))
// 	// Then, check that the role is either "admin" or "teacher".
// 	surveys.Use(middlewares.RoleMiddleware("admin", "teacher"))

// 	// Now define the survey endpoints.
// 	surveys.POST("/", surveyController.CreateSurvey)
// 	// surveys.GET("/", surveyController.GetAllSurveys)
// 	surveys.GET("/:id", surveyController.GetSurveyByID)
// 	surveys.PUT("/:id", surveyController.UpdateSurvey)
// 	surveys.GET("/search", surveyController.SearchSurveys)
// 	surveys.DELETE("/:id", surveyController.DeleteSurvey)
// }
