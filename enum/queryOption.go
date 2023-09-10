package enum

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type QueryOptionClosure func(c *gin.Context, GetRequestOptions GetRequestOptions) QueryOption

// QueryOption is a function that can be used to construct a query.
type QueryOption func(tx *gorm.DB) *gorm.DB
