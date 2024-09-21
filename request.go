package afast

import "net/http"

type Request struct {
	Method string
	Path   string
	Params map[string]interface{}
	Query  map[string][]string
}

func (r *Request) Param(name string) string {
	return r.Params[name].(string)
}

func (r *Request) ParamInt(name string) int {
	return r.Params[name].(int)
}

func NewRequest(r *http.Request, params *map[string]interface{}) *Request {
	return &Request{
		Method: r.Method,
		Path:   r.URL.Path,
		Params: *params,
		Query:  r.URL.Query(),
	}
}
