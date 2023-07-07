package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/tqrj/cd/enum"
	"github.com/tqrj/cd/service"
)

// CreateHandler handles
//
//	POST /T
//
// creates a new model T, responds with the created model T if successful.
//
// Request body:
//   - {...}  // fields of the model T
//
// Response:
//   - 200 OK: { T: {...} }
//   - 400 Bad Request: { error: "request band failed" }
//   - 422 Unprocessable Entity: { error: "create process failed" }
func CreateHandler[T any](opt *enum.CreateOption) gin.HandlerFunc {
	return func(c *gin.Context) {
		var model T
		if err := c.ShouldBindJSON(&model); err != nil {
			logger.WithContext(c).WithError(err).
				Warn("CreateHandler: Bind failed")
			ResponseError(c, CodeBadRequest, err)
			return
		}
		if opt.Pretreat != nil {
			res, err := opt.Pretreat(c, model)
			if err != nil {
				logger.WithContext(c).WithError(err).
					Warn("GetListHandler:Pretreat err")
				ResponseError(c, CodeBadRequest, err)
				return
			}
			model = res.(T)
		}
		logger.WithContext(c).Tracef("CreateHandler: Create %#v", model)
		err := service.Create(c, &model, opt, service.IfNotExist())
		if err != nil {
			logger.WithContext(c).WithError(err).
				Warn("CreateHandler: Create failed")
			ResponseError(c, CodeProcessFailed, err)
			return
		}
		c.JSON(200, SuccessResponseBody(model))
	}
}
