package main

import (
	"pathshala/config"
	"pathshala/middlewares"
	"pathshala/migrations"
	"pathshala/routes"
	"pathshala/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.Use(middlewares.PrometheusMiddleware())

	// Prometheus metrics endpoint
	r.GET("/metrics", middlewares.MetricsHandler())

	// Initialize DB
	config.ConnectDB()

	//Initialize REDIS
	config.InitRedis()

	// 3. Initialize Services
	// surveyService := services.NewSurveyService(config.DB)
	// reportService := services.NewReportService(config.DB)

	// Migration
	migrations.MigrateQuestions()

	utils.SeedAdminUser()

	// Register Routes
	routes.SetupAuthRoutes(r)
	routes.SetupUserRoutes(r, config.DB)
	routes.SetupProfileRoutes(r)
	routes.SetupCollegeRoutes(r, config.DB)
	routes.SetupCollegeTypeRoutes(r, config.DB)
	routes.SetupCategoryRoutes(r)
	routes.SetupMacroCategoryRoutes(r)
	routes.SetupHomeRouter(r)
	routes.SetupQuestionRoutes(r)
	routes.SetupTestRoutes(r)
	routes.SetupAnswerRoutes(r)
	routes.SetupResultRoutes(r)
	// routes.SetupReportRoutes(r, controllers.NewReportController(reportService))
	// routes.SetupSurveyRoutes()

	r.Run(":8080")
}
