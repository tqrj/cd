package controller

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/tqrj/cd/enum"
	"github.com/tqrj/cd/orm"
	"github.com/tqrj/cd/service"
	"reflect"
)

// GetListHandler handles
//
//	GET /T
//
// It returns a list of models.
//
// QueryOptions (See GetRequestOptions for more details):
//
//	limit, offset, order_by, desc, filter_by, filter_value, preload, total.
//
// Response:
//   - 200 OK: { Ts: [{...}, ...] }
//   - 400 Bad Request: { error: "request band failed" }
//   - 422 Unprocessable Entity: { error: "get process failed" }
func GetListHandler[T any](opt *enum.ListOption) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request enum.GetRequestOptions
		if err := c.ShouldBind(&request); err != nil {
			logger.WithContext(c).WithError(err).
				Warn("GetListHandler: bind request failed")
			ResponseError(c, CodeBadRequest, err)
			return
		}
		if opt.Pretreat != nil {
			var err error
			request, err = opt.Pretreat(request)
			if err != nil {
				logger.WithContext(c).WithError(err).
					Warn("GetListHandler:Pretreat err")
				ResponseError(c, CodeBadRequest, err)
				return
			}
		}

		request.Filters = c.QueryMap("filters")
		options := buildQueryOptions(request, opt.LimitMax, opt.Omit)
		options = append(options, opt.QueryOption)
		var dest []*T
		err := service.GetMany[T](c, &dest, options...)
		if err != nil {
			logger.WithContext(c).WithError(err).
				Warn("GetListHandler: GetMany failed")
			ResponseError(c, CodeProcessFailed, err)
			return
		}

		var addition []gin.H
		if request.Total {
			total, err := getCount[T](c, request.Filters, request.FiltersAt, opt.QueryOption)
			if err != nil {
				logger.WithContext(c).WithError(err).
					Warn("GetListHandler: getCount failed")
				addition = append(addition, gin.H{"totalError": err.Error()})
			} else {
				addition = append(addition, gin.H{"total": total})
			}
		}
		ResponseSuccess(c, dest, addition...)
	}
}

// GetByIDHandler handles
//
//	GET /T/:idParam
//
// QueryOptions (See GetRequestOptions for more details): preload
//
// Response:
//   - 200 OK: { T: {...} }
//   - 400 Bad Request: { error: "request band failed" }
//   - 422 Unprocessable Entity: { error: "get process failed" }
func GetByIDHandler[T orm.Model](idParam string, opt *enum.GetOption) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request enum.GetRequestOptions
		if err := c.ShouldBind(&request); err != nil {
			logger.WithContext(c).WithError(err).
				Warn("GetByIDHandler: bind request failed")
			ResponseError(c, CodeBadRequest, err)
			return
		}
		if opt.Pretreat != nil {
			var err error
			request, err = opt.Pretreat(request)
			if err != nil {
				logger.WithContext(c).WithError(err).
					Warn("GetListHandler:Pretreat err")
				ResponseError(c, CodeBadRequest, err)
				return
			}
		}
		request.Filters = c.QueryMap("filters")
		options := buildQueryOptions(request, 1, opt.Omit)
		options = append(options, opt.QueryOption)
		dest, err := getModelByID[T](c, idParam, options...)
		if err != nil {
			logger.WithContext(c).WithError(err).
				Warn("GetByIDHandler: getModelByID failed")
			ResponseError(c, CodeProcessFailed, err)
			return
		}
		ResponseSuccess(c, dest)
	}
}

// GetFieldHandler handles
//
//	GET /T/:idParam/field
//
// QueryOptions (See GetRequestOptions for more details):
//
//	limit, offset, order_by, desc, filter_by, filter_value, preload, total.
//
// Notice, all GetRequestOptions will be conditions for the field, for example:
//
//	GET /user/123/order?preload=Product
//
// Preloads User.Order.Product instead of User.Product.
//
// Response:
//   - 200 OK: { Fs: [{...}, ...] }  // field models
//   - 400 Bad Request: { error: "request band failed" }
//   - 422 Unprocessable Entity: { error: "get process failed" }
func GetFieldHandler[T orm.Model](idParam string, field string, opt *enum.GetOption) gin.HandlerFunc {
	field = nameToField(field, *new(T))

	return func(c *gin.Context) {
		var request enum.GetRequestOptions
		if err := c.ShouldBind(&request); err != nil {
			logger.WithContext(c).WithError(err).
				Warn("GetFieldHandler: bind request failed")
			ResponseError(c, CodeBadRequest, err)
			return
		}
		request.Filters = c.QueryMap("filters")
		options := buildQueryOptions(request, 1, opt.Omit)
		options = append(options, opt.QueryOption)
		model, err := getModelByID[T](c, idParam, service.Preload(field, options...))
		if err != nil {
			logger.WithContext(c).WithError(err).
				Warn("GetFieldHandler: getModelByID failed")
			ResponseError(c, CodeProcessFailed, err)
			return
		}

		fieldValue := reflect.ValueOf(model).
			Elem(). // because model is a pointer
			FieldByName(field)

		var addition []gin.H
		if request.Total && fieldValue.Kind() == reflect.Slice {
			total, err := getAssociationCount(c, model, field, request.Filters, request.FiltersAt, opt.QueryOption)
			if err != nil {
				logger.WithContext(c).WithError(err).
					Warn("GetFieldHandler: getAssociationCount failed")
				addition = append(addition, gin.H{"totalError": err.Error()})
			} else {
				addition = append(addition, gin.H{"total": total})
			}
		}

		ResponseSuccess(c, fieldValue.Interface(), addition...)
	}
}

func buildQueryOptions(request enum.GetRequestOptions, LimitMax int, omit []string) []enum.QueryOption {
	var options []enum.QueryOption
	if request.Limit > 0 && request.Limit <= LimitMax {
		options = append(options, service.WithPage(request.Limit, request.Offset))
	} else {
		options = append(options, service.WithPage(LimitMax, request.Offset))
	}
	if omit != nil && len(omit) != 0 {
		options = append(options, service.Omit(omit))
	}

	if request.OrderBy != "" {
		options = append(options, service.OrderBy(request.OrderBy, request.Descending))
	}

	for FilterBy, FilterValue := range request.Filters {
		if FilterBy != "" && FilterValue != "" {
			options = append(options, service.FilterBy(FilterBy, FilterValue))
		}
	}

	if len(request.FiltersAt) == 2 {
		options = append(options, service.FilterAt(request.FiltersAt))
	}

	for _, field := range request.Preload {
		// logger.WithField("field", field).Debug("Preload field")
		if field == "" {
			continue
		}

		options = append(options, service.Preload(field))
	}
	return options
}

// getModelByID gets idParam from url and get model from database
func getModelByID[T orm.Model](c *gin.Context, idParam string, options ...enum.QueryOption) (*T, error) {
	var model T

	id := c.Param(idParam)
	if id == "" {
		logger.WithContext(c).WithField("idParam", idParam).
			Warn("getModelByID: id is empty")
		return &model, ErrMissingID
	}

	err := service.GetByID[T](c, id, &model, options...)
	return &model, err
}

func getCount[T any](ctx context.Context, filters map[string]string, filterAt []string, option enum.QueryOption) (total int64, err error) {
	var options []enum.QueryOption
	for filterBy, filterValue := range filters {
		if filterBy != "" && filterValue != "" {
			options = append(options, service.FilterBy(filterBy, filterValue))
		}
	}
	if len(filterAt) == 2 {
		options = append(options, service.FilterAt(filterAt))
	}
	if option != nil {
		options = append(options, option)
	}
	total, err = service.Count[T](ctx, options...)
	return total, err
}

func getAssociationCount(ctx context.Context, model any, field string, filters map[string]string, filterAt []string, option enum.QueryOption) (total int64, err error) {
	var options []enum.QueryOption

	for filterBy, filterValue := range filters {
		if filterBy != "" && filterValue != "" {
			options = append(options, service.FilterBy(filterBy, filterValue))
		}
	}

	if len(filterAt) == 2 {
		options = append(options, service.FilterAt(filterAt))
	}
	if option != nil {
		options = append(options, option)
	}
	count, err := service.CountAssociations(ctx, model, field, options...)
	return count, err
}
