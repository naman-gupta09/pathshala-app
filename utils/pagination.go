// utils/pagination.go
package utils

import (
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Pagination struct {
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	Offset     int         `json:"offset"`
	TotalCount int         `json:"total_count"`
	TotalPages int         `json:"total_pages"`
	Data       interface{} `json:"data"`
}

func GetPaginationParams(c *gin.Context) (int, int) {
	const (
		defaultPage  = 1
		defaultLimit = 10
		maxLimit     = 100
	)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = defaultPage
	}
	if limit < 1 {
		limit = defaultLimit
	} else if limit > maxLimit {
		limit = maxLimit
	}

	return page, limit
}

// PaginateSlice paginates an array/slice in-memory
func PaginateSlice[T any](c *gin.Context, fullData []T) {
	page, limit := GetPaginationParams(c)
	total := len(fullData)

	offset := (page - 1) * limit
	end := offset + limit
	if end > total {
		end = total
	}

	var pagedData []T
	if offset < total {
		pagedData = fullData[offset:end]
	} else {
		pagedData = []T{}
	}

	response := Pagination{
		Page:       page,
		Limit:      limit,
		Offset:     offset,
		TotalCount: total,
		TotalPages: int(math.Ceil(float64(total) / float64(limit))),
		Data:       pagedData,
	}

	c.JSON(http.StatusOK, response)
}

type PaginationParams struct {
	Page   int
	Limit  int
	Offset int
}

func GetPaginationParamsWithOffset(c *gin.Context) PaginationParams {
	page, limit := GetPaginationParams(c)
	offset := (page - 1) * limit
	return PaginationParams{
		Page:   page,
		Limit:  limit,
		Offset: offset,
	}
}

func SendPaginatedResponse(c *gin.Context, pagination PaginationParams, totalCount int64, data interface{}) {
	totalPages := int(math.Ceil(float64(totalCount) / float64(pagination.Limit)))
	c.JSON(http.StatusOK, Pagination{
		Page:       pagination.Page,
		Limit:      pagination.Limit,
		Offset:     pagination.Offset,
		TotalCount: int(totalCount),
		TotalPages: totalPages,
		Data:       data,
	})
}
