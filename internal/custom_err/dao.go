package custom_err

import (
	"github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

var (
	ErrInitDaoFailed  = errors.New("init dao failed")
	ErrDuplicateEntry = mysql.MySQLError{
		Number: 1062,
	}

	// ErrRecordNotFound no records found
	ErrRecordNotFound = gorm.ErrRecordNotFound
)
