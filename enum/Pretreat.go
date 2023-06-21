package enum

import "github.com/gin-gonic/gin"

// Pretreat 更新和新增
type Pretreat func(c *gin.Context, model any) (any, error)
type DeletePretreat func(c *gin.Context, id string) (string, error)

// GetPretreat list和get
type GetPretreat func(c *gin.Context, GetRequestOptions GetRequestOptions) (GetRequestOptions, error)
