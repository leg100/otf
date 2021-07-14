package ots

import (
	"errors"

	"gorm.io/gorm"
)

func IsNotFound(err error) bool {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return true
	}
	return false
}
