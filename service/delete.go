package service

import (
	"context"
	"github.com/tqrj/cd/enum"
	"github.com/tqrj/cd/orm"
)

// Delete a model from database.
func Delete(ctx context.Context, model any) (rowsAffected int64, err error) {
	logger.WithContext(ctx).
		WithField("model", model).Trace("Delete model")
	result := orm.DB.WithContext(ctx).Delete(model)
	return result.RowsAffected, result.Error
}

// DeleteByID deletes a model from database by its ID.
func DeleteByID[T orm.Model](ctx context.Context, id any, opt *enum.DelOption) (rowsAffected int64, err error) {
	logger.WithContext(ctx).
		WithField("id", id).
		Trace("DeleteByID: Delete model by ID")

	var model T
	if err := GetByID[T](ctx, id, &model); err != nil {
		logger.WithContext(ctx).
			WithField("id", id).WithError(err).
			Warn("DeleteByID: GetByID failed")
		return 0, err
	}
	db := orm.DB.WithContext(ctx)
	if opt.QueryOption != nil {
		db = opt.QueryOption(db)
	}
	result := db.Delete(&model)
	if result.Error != nil {
		logger.WithContext(ctx).
			WithError(result.Error).Warn("DeleteByID: failed")
	}
	return result.RowsAffected, result.Error
}
