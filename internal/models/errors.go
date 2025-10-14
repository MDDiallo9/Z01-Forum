package models

import (
	"errors"
)

var (
	ErrNoRecords = errors.New("models: no matching record found")
	ErrDuplicateRecord = errors.New("models: duplicate record")
)