package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/tqrj/cd/enum"
	"github.com/tqrj/cd/log"
	"github.com/tqrj/cd/orm"
	"github.com/tqrj/cd/service"
)

// UpdateHandler handles
//
//	PUT /T/:idParam
//
// Updates the model T with the given id.
//
// Request body:
//   - {"field": "new_value", ...}   // fields to update
//
// Response:
//   - 200 OK: { updated: true }
//   - 400 Bad Request: { error: "missing id or bind fields failed" }
//   - 404 Not Found: { error: "record with id not found" }
//   - 422 Unprocessable Entity: { error: "update process failed" }
func UpdateHandler[T orm.Model](idParam string, opt *enum.UpdateOption) gin.HandlerFunc {
	return func(c *gin.Context) {
		var model T

		id := c.Param(idParam) // NOTICE: id is a string
		if id == "" {
			logger.WithContext(c).WithField("idParam", idParam).
				Warn("UpdateHandler: Missing id")
			ResponseError(c, CodeBadRequest, ErrMissingID)
			return
		}

		if err := service.GetByID[T](c, id, &model); err != nil {
			logger.WithContext(c).WithError(err).
				Warn("UpdateHandler: GetByID failed")
			ResponseError(c, CodeNotFound, err)
			return
		}

		var updatedModel = model
		if err := c.ShouldBindJSON(&updatedModel); err != nil {
			logger.WithContext(c).WithError(err).
				Warn("UpdateHandler: Bind failed")
			ResponseError(c, CodeBadRequest, err)
			return
		}
		if opt.Pretreat != nil {
			res, err := opt.Pretreat(c, updatedModel)
			if err != nil {
				logger.WithContext(c).WithError(err).
					Warn("GetListHandler:Pretreat err")
				ResponseError(c, CodeBadRequest, err)
				return
			}
			updatedModel = res.(T)
		}

		log.Logger.Tracef("UpdateHandler: Update %#v, id=%v", updatedModel, id)

		_, oldID := model.Identity()
		_, newID := updatedModel.Identity()
		if oldID != newID {
			logger.WithContext(c).WithField("idParam", idParam).
				WithField("oldID", oldID).
				WithField("newID", newID).
				Warn("UpdateHandler: id mismatch: cannot update id")
			ResponseError(c, CodeBadRequest, ErrUpdateID)
			return
		}

		_, err := service.Update(c, &updatedModel, opt)
		if err != nil {
			logger.WithContext(c).WithError(err).
				Warn("UpdateHandler: Update failed")
			ResponseError(c, CodeProcessFailed, err)
			return
		}
		ResponseSuccess(c, &updatedModel)
	}
}
