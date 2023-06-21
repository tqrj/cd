package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/tqrj/cd/enum"
	"github.com/tqrj/cd/orm"
)

// Update all fields of an existing model in database.
func Update(ctx context.Context, model any, opt *enum.UpdateOption) (rowsAffected int64, err error) {
	logger.WithContext(ctx).
		WithField("model", model).Trace("Update model")

	if model == nil {
		logger.WithContext(ctx).
			Warn("Update: model is nil, nothing to update")
		return 0, ErrNoRecord
	}
	db := orm.DB.WithContext(ctx)
	db = Omit(opt.Omit)(db)
	if opt.QueryOption != nil {
		db = opt.QueryOption(db)
	}
	result := db.Save(model)
	if result.Error != nil {
		logger.WithContext(ctx).
			WithError(result.Error).Warn("Update: failed")
	}
	return result.RowsAffected, result.Error
}

var (
	ErrNoRecord        = errors.New("no record found")
	ErrMultipleRecords = errors.New("multiple records found")
)

// UpdateField updates a single fields of an existing model in database.
// It will try to GetByID first, to make sure the model exists, before updating.
func UpdateField[T orm.Model](ctx context.Context, id any, field string, value interface{}) (rowsAffected int64, err error) {
	logger.WithContext(ctx).
		WithField("model", fmt.Sprintf("%T", *new(T))).
		WithField("id", id).WithField("field", field).
		WithField("value", value).Trace("UpdateField")

	var record T
	if err := GetByID[T](ctx, id, &record); err != nil {
		logger.WithContext(ctx).
			WithField("id", id).WithError(err).
			Warn("UpdateField: GetByID failed")
		return 0, err
	}
	result := orm.DB.WithContext(ctx).Model(&record).Update(field, value)
	if result.Error != nil {
		logger.WithContext(ctx).
			WithError(result.Error).Warn("UpdateField: failed")
	}
	return result.RowsAffected, result.Error
}
