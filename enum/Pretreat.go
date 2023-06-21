package enum

// Pretreat 更新和新增
type Pretreat func(model any) (any, error)
type DeletePretreat func(id string) (string, error)

// GetPretreat list和get
type GetPretreat func(GetRequestOptions GetRequestOptions) (GetRequestOptions, error)
