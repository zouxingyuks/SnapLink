package custom_err

import (
	"gorm.io/gorm"
)

var (

	// ErrRecordNotFound no records found
	ErrRecordNotFound = gorm.ErrRecordNotFound
)
