package controllers

import (
	"net/http"
	"pathshala/models"
	"pathshala/utils"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

func CreateStudent(c *gin.Context, db *gorm.DB) {
	var input struct {
		Name           string  `json:"name" binding:"required"`
		Email          string  `json:"email" binding:"required,email"`
		Password       string  `json:"password" binding:"required,min=6"`
		SecondaryEmail *string `json:"secondary_email,omitempty" binding:"omitempty,email"`
		CollegeName    string  `json:"college_name" binding:"required"`
		Branch         string  `json:"branch" binding:"required"`
		Gender         string  `json:"gender" binding:"required,oneof=male female"`
		Status         string  `json:"status" binding:"required,oneof=active inactive"` // ✅ Add this
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors := utils.FormatValidationError(validationErrors)
			c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 🔍 Find College ID from College Name
	var college models.College
	if err := db.Where("name = ?", input.CollegeName).First(&college).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid college name"})
		return
	}

	//  Hash the password
	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	//  Create User
	user := models.User{
		Name:           input.Name,
		Email:          input.Email,
		Password:       hashedPassword,
		SecondaryEmail: input.SecondaryEmail,
		CollegeID:      &college.ID,
		Role:           "student",
	}

	if err := db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create student user"})
		return
	}

	//  Create Student record
	student := models.Student{
		UserID: user.ID,
		Branch: input.Branch,
		Gender: input.Gender,
		Status: input.Status,
	}

	if err := db.Create(&student).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create student details"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Student created successfully",
		"user_id": user.ID,
	})
}

func CreateTeacher(c *gin.Context, db *gorm.DB) {
	requesterRole, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	var input struct {
		Name         string `json:"name" binding:"required"`
		Email        string `json:"email" binding:"required,email"`
		Password     string `json:"password" binding:"required"`
		State        string `json:"state" binding:"required"`
		CollegeName  string `json:"college_name" binding:"required"`
		SuperTeacher bool   `json:"super_teacher"`
		TeacherType  string `json:"teacher_type" binding:"required"`
		Status       string `json:"status" binding:"required,oneof=active inactive"` // ✅ Add this

	}

	if err := c.ShouldBindJSON(&input); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors := utils.FormatValidationError(validationErrors)
			c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//  Find college by name
	var college models.College
	if err := db.Where("name = ?", input.CollegeName).First(&college).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "College not found"})
		return
	}

	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	//  Create user with found college ID
	user := models.User{
		Name:      input.Name,
		Email:     input.Email,
		Password:  hashedPassword,
		Role:      "teacher",
		CollegeID: &college.ID,
	}

	if requesterRole == "teacher" && user.Role == "teacher" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Teachers can only create students"})
		return
	}
	result := db.Create(&user)
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "duplicate key value violates unique constraint") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email already exists"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		}
		return
	}

	//  Create teacher profile
	teacher := models.Teacher{
		UserID:      user.ID,
		State:       input.State,
		TeacherType: input.TeacherType,
		Super:       input.SuperTeacher,
		Status:      input.Status,
	}

	if err := db.Create(&teacher).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create teacher profile"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":       "Teacher created successfully",
		"user_id":       user.ID,
		"name":          user.Name,
		"email":         user.Email,
		"college":       input.CollegeName,
		"teacher_type":  input.TeacherType,
		"state":         input.State,
		"super_teacher": input.SuperTeacher,
	})

}

func GetUsersByRole(c *gin.Context, db *gorm.DB, role string) {
	column := c.Query("column")
	value := c.Query("value")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	if role == "student" {
		var results []models.StudentResponse
		query := db.Table("users").
			Select("users.id, users.name, users.email, users.college_id, COALESCE(colleges.name, 'N/A') as college, users.role, students.status").
			Joins("LEFT JOIN students ON users.id = students.user_id").
			Joins("LEFT JOIN colleges ON users.college_id = colleges.id").
			Where("users.role = ?", role)

		// Apply search filter if provided
		if column != "" && value != "" {
			validColumns := map[string]string{
				"name":    "users.name",
				"email":   "users.email",
				"college": "colleges.name",
				"status":  "students.status",
			}

			dbColumn, ok := validColumns[column]
			if !ok {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid search column"})
				return
			}
			query = query.Where(dbColumn+" ILIKE ?", "%"+value+"%")
		}

		// Apply pagination
		err := query.Offset(offset).Limit(limit).Scan(&results).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}
		if len(results) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "No students found"})
			return
		}
		// c.JSON(http.StatusOK, results)
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Users fetched successfully",
			"data":    results,
		})

		return
	}

	if role == "teacher" {
		var results []models.TeacherResponse
		query := db.Table("users").
			Select("users.id, users.name, users.email, users.college_id, COALESCE(colleges.name, 'N/A') as college, users.role, teachers.state, teachers.teacher_type").
			Joins("LEFT JOIN teachers ON users.id = teachers.user_id").
			Joins("LEFT JOIN colleges ON users.college_id = colleges.id").
			Where("users.role = ?", role)

		// Apply search filter if provided
		if column != "" && value != "" {
			validColumns := map[string]string{
				"name":         "users.name",
				"email":        "users.email",
				"college":      "colleges.name",
				"state":        "teachers.state",
				"teacher_type": "teachers.teacher_type",
			}

			dbColumn, ok := validColumns[column]
			if !ok {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid search column"})
				return
			}
			query = query.Where(dbColumn+" ILIKE ?", "%"+value+"%")
		}

		// Apply pagination
		err := query.Offset(offset).Limit(limit).Scan(&results).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}
		if len(results) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "No teachers found"})
			return
		}
		// c.JSON(http.StatusOK, results)
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Users fetched successfully",
			"data":    results,
		})

		return
	}

	// fallback
	c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported role"})
}

// After changes in user struct
func UpdateUser(c *gin.Context, db *gorm.DB) {
	id := c.Param("id")

	var existingUser models.User
	if err := db.First(&existingUser, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var input struct {
		Name        string `json:"name"`
		Email       string `json:"email"`
		CollegeID   *uint  `json:"college_id"`
		Role        string `json:"role"`
		Status      string `json:"status"`       // For student
		State       string `json:"state"`        // For teacher
		TeacherType string `json:"teacher_type"` // For teacher
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Update User fields
	existingUser.Name = input.Name
	existingUser.Email = input.Email
	existingUser.CollegeID = input.CollegeID
	existingUser.Role = input.Role

	if err := db.Save(&existingUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	// Update role-specific fields
	switch input.Role {
	case "student":
		db.Model(&models.Student{}).Where("user_id = ?", existingUser.ID).Updates(models.Student{Status: input.Status})
	case "teacher":
		db.Model(&models.Teacher{}).Where("user_id = ?", existingUser.ID).Updates(models.Teacher{
			State:       input.State,
			TeacherType: input.TeacherType,
		})
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

func DeleteUser(c *gin.Context, db *gorm.DB) {
	var user models.User
	if err := db.First(&user, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	switch user.Role {
	case "student":
		db.Where("user_id = ?", user.ID).Delete(&models.Student{})
	case "teacher":
		db.Where("user_id = ?", user.ID).Delete(&models.Teacher{})
	}

	db.Delete(&user)
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
