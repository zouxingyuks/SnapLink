package dao

import (
	"github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
)

var (
	ErrInitDaoFailed  = errors.New("init dao failed")
	ErrMarshalType    = errors.New("ErrMarshalType")
	ErrUnmarshalType  = errors.New("ErrUnmarshalType")
	ErrDuplicateEntry = mysql.MySQLError{
		Number: 1062,
	}
)
