package utils

import (
	"fmt"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

func ApplySearchFilter(query *gorm.DB, searchColumn, searchValue string) *gorm.DB {
	if searchColumn == "" || searchValue == "" {
		return query
	}

	// Try integer comparison
	if num, err := strconv.Atoi(searchValue); err == nil {
		query = query.Where(fmt.Sprintf("%s = ?", searchColumn), num)
		return query
	}

	// Try float comparison
	if fnum, err := strconv.ParseFloat(searchValue, 64); err == nil {
		query = query.Where(fmt.Sprintf("%s = ?", searchColumn), fnum)
		return query
	}

	// Default to case-insensitive string LIKE
	query = query.Where(fmt.Sprintf("LOWER(%s) LIKE ?", searchColumn), "%"+strings.ToLower(searchValue)+"%")
	return query
}

// ApplyDynamicFilters applies filters based on field types (text, int, float)
/*func Searchfilter(query *gorm.DB, filters map[string][]string, allowedFields map[string]string) *gorm.DB {
	for key, values := range filters {
		if len(values) == 0 {
			continue
		}
		fieldType, allowed := allowedFields[key]
		if !allowed {
			continue
		}

		value := values[0]

		switch fieldType {
		case "string":
			// Case-insensitive LIKE search
			query = query.Where(fmt.Sprintf("LOWER(%s) LIKE ?", key), "%"+strings.ToLower(value)+"%")

		case "int":
			var num int
			if _, err := fmt.Sscanf(value, "%d", &num); err == nil {
				query = query.Where(fmt.Sprintf("%s = ?", key), num)
			}

		case "float":
			var num float64
			if _, err := fmt.Sscanf(value, "%f", &num); err == nil {
				query = query.Where(fmt.Sprintf("%s = ?", key), num)
			}
		}
	}
	return query
}
*/
