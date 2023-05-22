package router

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/tqrj/cd/controller"
	"github.com/tqrj/cd/enum"
	"github.com/tqrj/cd/orm"
	"reflect"
)

// Crud add a group of CRUD routes for model T to the base router
// on relativePath. For example, if the base router handles route to
//
//	/base
//
// then the CRUD routes will be:
//
//	GET|POST|PUT|DELETE /base/relativePath
//
// where relativePath is recommended to be the plural form of the model name.
//
// Example:
//
//	r := gin.Default()
//	Crud[User](r, "/users")
//
// adds the following routes:
//
//	   GET /users/
//	   GET /users/:UserId
//	  POST /users/
//	   PUT /users/:UserId
//	DELETE /users/:UserId
//
// and with options parameters, it's optional to add the following routes:
//   - GetNested()    =>    GET /users/:UserId/friends
//   - CreateNested() =>   POST /users/:UserId/friends
//   - DeleteNested() => DELETE /users/:UserId/friends/:FriendId
func Crud[T orm.Model](base gin.IRouter, relativePath string, opt *enum.CurdOption, crudGroups ...enum.CrudGroup) gin.IRouter {
	group := base.Group(relativePath)

	if !gin.IsDebugging() { // GIN_MODE == "release"
		logger.WithField("model", getTypeName[T]()).
			//WithField("basePath", base). // we cannot get the base path from gin.IRouter
			WithField("relativePath", relativePath).
			Info("Crud: Adding CRUD routes for model")
	}

	crudGroups = append(crudGroups, crud[T](opt))

	for _, option := range crudGroups {
		group = option(group)
	}

	return group
}

func DefaultCurdOption() *enum.CurdOption {
	return &enum.CurdOption{
		ListOption: enum.ListOption{
			Enable:   true,
			Omit:     nil,
			LimitMax: 10,
		},
		GetOption: enum.GetOption{
			Enable: true,
			Omit:   nil,
		},
		UpdateOption: enum.UpdateOption{Enable: true},
		CreateOption: enum.CreateOption{Enable: true},
		DelOption:    enum.DelOption{Enable: true},
	}
}

// crud add CRUD routes for model T to the group:
//
//	   GET /
//	   GET /:idParam
//	  POST /
//	   PUT /:idParam
//	DELETE /:idParam
func crud[T orm.Model](opt *enum.CurdOption) enum.CrudGroup {
	idParam := getIdParam[T]()
	return func(group *gin.RouterGroup) *gin.RouterGroup {
		if opt.ListOption.Enable {
			group.GET("", controller.GetListHandler[T](&opt.ListOption))
		}
		if opt.GetOption.Enable {
			group.GET(fmt.Sprintf("/:%s", idParam), controller.GetByIDHandler[T](idParam, &opt.GetOption))
		}
		if opt.CreateOption.Enable {
			group.POST("", controller.CreateHandler[T]())
		}
		if opt.UpdateOption.Enable {
			group.PUT(fmt.Sprintf("/:%s", idParam), controller.UpdateHandler[T](idParam))
		}
		if opt.DelOption.Enable {
			group.DELETE(fmt.Sprintf("/:%s", idParam), controller.DeleteHandler[T](idParam))
		}

		return group
	}
}

// GetNested add a GET route to the group for querying a nested model:
//
//	GET /:parentIdParam/field
func GetNested[P orm.Model, N orm.Model](field string, opt *enum.GetOption) enum.CrudGroup {
	parentIdParam := getIdParam[P]()
	return func(group *gin.RouterGroup) *gin.RouterGroup {
		relativePath := fmt.Sprintf("/:%s/%s", parentIdParam, field)

		if !gin.IsDebugging() { // GIN_MODE == "release"
			logger.WithField("parent", getTypeName[P]()).
				WithField("child", getTypeName[N]()).
				WithField("relativePath", relativePath).
				Info("Crud: Adding GET route for getting nested model")
		}

		group.GET(relativePath,
			controller.GetFieldHandler[P](parentIdParam, field, opt),
		)
		// there is no GET /:parentIdParam/:field/:childIdParam,
		// because it is equivalent to GET /:childModel/:childIdParam.
		// So there is also no PUT /:parentIdParam/:field/:childIdParam.
		// It is verbose and unnecessary.
		return group
	}
}

// CreateNested add a POST route to the group for creating a nested model:
//
//	POST /:parentIdParam/field
func CreateNested[P orm.Model, N orm.Model](field string) enum.CrudGroup {
	parentIdParam := getIdParam[P]()
	return func(group *gin.RouterGroup) *gin.RouterGroup {
		relativePath := fmt.Sprintf("/:%s/%s", parentIdParam, field)

		if !gin.IsDebugging() { // GIN_MODE == "release"
			logger.WithField("parent", getTypeName[P]()).
				WithField("child", getTypeName[N]()).
				WithField("relativePath", relativePath).
				Info("Crud: Adding POST route for creating nested model")
		}

		group.POST(relativePath,
			controller.CreateNestedHandler[P, N](parentIdParam, field),
		)
		return group
	}
}

// DeleteNested add a DELETE route to the group for deleting a nested model:
//
//	DELETE /:parentIdParam/field/:childIdParam
func DeleteNested[P orm.Model, T orm.Model](field string) enum.CrudGroup {
	parentIdParam := getIdParam[P]()
	childIdParam := getIdParam[T]()
	return func(group *gin.RouterGroup) *gin.RouterGroup {
		relativePath := fmt.Sprintf("/:%s/%s/:%s", parentIdParam, field, childIdParam)

		if !gin.IsDebugging() { // GIN_MODE == "release"
			logger.WithField("parent", getTypeName[P]()).
				WithField("child", getTypeName[T]()).
				WithField("relativePath", relativePath).
				Info("Crud: Adding DELETE route for deleting nested model")
		}

		group.DELETE(relativePath,
			controller.DeleteNestedHandler[P, T](parentIdParam, field, childIdParam),
		)
		return group
	}
}

// CrudNested = GetNested + CreateNested + DeleteNested
func CrudNested[P orm.Model, T orm.Model](field string, opt *enum.CurdOption) enum.CrudGroup {
	return func(group *gin.RouterGroup) *gin.RouterGroup {

		if opt.GetOption.Enable {
			group = GetNested[P, T](field, &opt.GetOption)(group)

		}

		if opt.CreateOption.Enable {
			group = CreateNested[P, T](field)(group)
		}
		if opt.DelOption.Enable {
			group = DeleteNested[P, T](field)(group)
		}
		return group
	}
}

// getIdParam Model => "ModelID"
func getIdParam[T orm.Model]() string {
	model := *new(T)
	modelName := reflect.TypeOf(model).Name()
	idField, _ := model.Identity()
	idParam := modelName + idField

	return idParam
}

// getTypeName is a helper function to get the type name of a generic type T.
func getTypeName[T any]() string {
	model := *new(T)
	return reflect.TypeOf(model).Name()
}
