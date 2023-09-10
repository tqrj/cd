package enum

import (
	"github.com/gin-gonic/gin"
)

type ListOption struct {
	Enable             bool
	Omit               []string
	LimitMax           int
	QueryOptionClosure QueryOptionClosure
	Pretreat           GetPretreat
}

type GetOption struct {
	Enable             bool
	Omit               []string
	QueryOptionClosure QueryOptionClosure
	Pretreat           GetPretreat
}

type UpdateOption struct {
	Enable   bool
	Omit     []string
	Pretreat Pretreat
	LimitID  []int64
}

type CreateOption struct {
	Enable   bool
	Omit     []string
	Pretreat Pretreat
}

type DelOption struct {
	Enable   bool
	Pretreat DeletePretreat
	LimitID  []int64
}

// CrudGroup is options to construct the router group.
//
// By adding GetNested, CreateNested, DeleteNested to Crud,
// you can add CRUD routes for a nested model (Parent.Child).
//
// Or use CrudNested to add all three options above.
type CrudGroup func(group *gin.RouterGroup) *gin.RouterGroup

type CurdOption struct {
	ListOption
	GetOption
	UpdateOption
	CreateOption
	DelOption
}
