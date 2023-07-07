package enum

import (
	"gorm.io/gorm"
)

// QueryOption is a function that can be used to construct a query.
type QueryOption func(tx *gorm.DB) *gorm.DB
