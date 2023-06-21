package enum

// GetRequestOptions is the query options (?opt=val) for GET requests:
//
//	limit=10&offset=4&                 # pagination
//	order_by=id&desc=true&             # ordering
//	filter_by=name&filter_value=John&  # filtering
//	total=true&                        # return total count (all available records under the filter, ignoring pagination)
//	preload=Product&preload=Product.Manufacturer  # preloading: loads nested models as well
//
// It is used in GetListHandler, GetByIDHandler and GetFieldHandler, to bind
// the query parameters in the GET request url.
type GetRequestOptions struct {
	Limit      int               `form:"limit"`
	Offset     int               `form:"offset"`
	OrderBy    string            `form:"order_by"`
	Descending bool              `form:"desc"`
	Filters    map[string]string `form:"filters"`
	FiltersAt  []string          `form:"filters_at"`
	Preload    []string          `form:"preload"` // fields to preload
	Total      bool              `form:"total"`   // return total count ?
}
