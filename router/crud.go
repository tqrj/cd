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

func DefaultCrudOption() *enum.CurdOption {
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
			group.POST("", controller.CreateHandler[T](&opt.CreateOption))
		}
		if opt.UpdateOption.Enable {
			group.PUT(fmt.Sprintf("/:%s", idParam), controller.UpdateHandler[T](idParam, &opt.UpdateOption))
		}
		if opt.DelOption.Enable {
			group.DELETE(fmt.Sprintf("/:%s", idParam), controller.DeleteHandler[T](idParam, &opt.DelOption))
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
