package routes

import (
	"pathshala/controllers"
	"pathshala/middlewares"
	"time"

	"github.com/gin-gonic/gin"
)

// SetupQuestionRoutes initializes question-related routes
func SetupQuestionRoutes(r *gin.Engine) {
	questionGroup := r.Group("/api/questions").Use(middlewares.AuthMiddleware(), middlewares.RoleMiddleware("admin", "teacher"))

	questionGroup.POST("/", controllers.AddQuestion)                                               // Add question
	questionGroup.PUT("/:id", controllers.EditQuestion)                                            // Edit question
	questionGroup.DELETE("/:id", controllers.DeleteQuestion)                                       // Delete question
	questionGroup.GET("/", controllers.GetQuestions, middlewares.TimeoutMiddleware(5*time.Second)) // Get and Search questions

	testQuestionGroup := r.Group("/api/tests").Use(middlewares.AuthMiddleware(), middlewares.RoleMiddleware("admin", "teacher"))

	testQuestionGroup.POST("/:test_id/addExstingQuestion", controllers.AddExistingQuestionToTest)                             // Add existing questions to test
	testQuestionGroup.POST("/:test_id/addNewQuestion", controllers.AddNewQuestionToTest)                                      // Add new question to test
	testQuestionGroup.DELETE("/:test_id/questions/:question_id", controllers.DeleteTestQuestion)                              // Delete a questions from test
	testQuestionGroup.GET("/:test_id/questions/", controllers.GetTestQuestions, middlewares.TimeoutMiddleware(5*time.Second)) // Get and Search questions from a test
	testQuestionGroup.PUT("/:test_id/questions/:question_id", controllers.EditQuestionOfTest)                                 // Edit questions from a test

}
