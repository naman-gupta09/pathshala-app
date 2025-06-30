package controllers

import (
	"net/http"
	"pathshala/services"

	"github.com/gin-gonic/gin"
)

type ReportController struct {
	Service services.ReportServiceInterface
}

func NewReportController(service services.ReportServiceInterface) *ReportController {
	return &ReportController{Service: service}
}

func (rc ReportController) GetReportTypes(c *gin.Context) {
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")

	if userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized access"})
		return
	}

	types, err := rc.Service.GetAllReportTypes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch report types"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"requested_by": userID,
		"types":        types,
	})
}

func (rc ReportController) GetTestStudentScores(c *gin.Context) {
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")

	if userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized access"})
		return
	}

	testID := c.Param("test_id") // No conversion, keep UUID as string

	data, err := rc.Service.GetTestWithStudentScores(testID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch test scores"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"requested_by": userID,
		"data":         data,
	})
}

func (rc ReportController) StudentParticipationRanking(c *gin.Context) {
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")

	if userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized access"})
		return
	}

	data, err := rc.Service.GetStudentParticipationRanking()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch participation ranking"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"requested_by": userID,
		"data":         data,
	})
}
