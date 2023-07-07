package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/tqrj/cd/enum"
	"github.com/tqrj/cd/orm"
	"github.com/tqrj/cd/service"
)

// DeleteHandler handles
//
//	DELETE /T/:idParam
//
// Deletes the model T with the given id.
//
// Request body: none
//
// Response:
//   - 200 OK: { deleted: true }
//   - 400 Bad Request: { error: "missing id" }
//   - 422 Unprocessable Entity: { error: "delete process failed" }
func DeleteHandler[T orm.Model](idParam string, opt *enum.DelOption) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param(idParam)
		if id == "" {
			logger.WithContext(c).
				WithField("idParam", idParam).
				Warn("DeleteHandler: read id param failed")
			ResponseError(c, CodeBadRequest, ErrMissingID)
			return
		}
		if Contains(opt.LimitID, cast.ToInt64(id)) {
			logger.WithContext(c).
				WithField("idParam", idParam).
				Warn("DeleteHandler: limit ID failed")
			ResponseError(c, CodeBadRequest, ErrMissingID)
			return
		}
		logger.WithContext(c).
			Tracef("DeleteHandler: Delete %T, id=%v", *new(T), id)
		if opt.Pretreat != nil {
			var err error
			id, err = opt.Pretreat(c, idParam)
			if err != nil {
				logger.WithContext(c).WithError(err).
					Warn("GetListHandler:Pretreat err")
				ResponseError(c, CodeBadRequest, err)
				return
			}
		}
		_, err := service.DeleteByID[T](c, id, opt)
		if err != nil {
			ResponseError(c, CodeProcessFailed, err)
			return
		}
		ResponseSuccess(c, nil, gin.H{"deleted": true})
	}
}

// Contains returns true if an element is present in a collection.
func Contains[T comparable](collection []T, element T) bool {
	for _, item := range collection {
		if item == element {
			return true
		}
	}

	return false
}
