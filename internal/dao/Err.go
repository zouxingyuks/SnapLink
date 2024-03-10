package dao

import (
	"github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
)

var (
	MarshalTypeError   = errors.New("MarshalTypeError")
	UnmarshalTypeError = errors.New("UnmarshalTypeError")
	DuplicateEntry     = mysql.MySQLError{
		Number: 1062,
	}
)
