package afast

import (
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"strings"
)

type handle struct {
	handle      AFastHandler
	middlewares []AFastMiddleware
}

type view struct {
	view        AFastView
	middlewares []AFastMiddleware
}

type ws struct {
	websocket   AFastWebsocket
	middlewares []AFastMiddleware
}

type Router struct {
	middlewares []AFastMiddleware
	router      *map[string]handle
	view        *view
	websocket   *ws
	children    *map[string]*Router
}

func NewRouter() *Router {
	return &Router{
		middlewares: []AFastMiddleware{},
		router:      &map[string]handle{},
		view:        nil,
		children:    &map[string]*Router{},
	}
}

func json(ro *Router, depth int, paths ...string) string {
	result := strings.Builder{}
	for method, h := range *ro.router {
		result.WriteString(strings.Repeat(" ", 4*depth))
		result.WriteString(method)
		result.WriteString(" /")
		result.WriteString(strings.Join(paths, "/"))

		funcValue := reflect.ValueOf(h.handle)
		funcName := runtime.FuncForPC(funcValue.Pointer()).Name()
		result.WriteString(" ")
		result.WriteString(funcName)

		result.WriteString("\n")
	}
	if ro.view != nil {
		result.WriteString(strings.Repeat(" ", 4*depth))
		result.WriteString("View /")
		result.WriteString(strings.Join(paths, "/"))

		viewValue := reflect.TypeOf(ro.view.view)
		structName := viewValue.Name()
		result.WriteString(" ")
		result.WriteString(structName)

		result.WriteString("\n")
	}
	if ro.websocket != nil {
		result.WriteString(strings.Repeat(" ", 4*depth))
		result.WriteString("Websocket /")
		result.WriteString(strings.Join(paths, "/"))

		viewValue := reflect.TypeOf(ro.websocket.websocket)
		structName := viewValue.Name()
		result.WriteString(" ")
		result.WriteString(structName)

		result.WriteString("\n")
	}
	if ro.children != nil {
		for path, child := range *ro.children {
			result.WriteString(strings.Repeat(" ", 4*depth))
			result.WriteString(path)
			result.WriteString("\n")
			result.WriteString(json(child, depth+1, append(paths, path)...))
		}
	}
	return result.String()
}

func (r *Router) Map() string {
	return json(r, 0)
}

func (r *Router) MapJson() string {
	return json(r, 0)
}

func (r *Router) Get(path string, handler AFastHandler, middlewares ...AFastMiddleware) {
	r.method("GET", path, handler, middlewares...)
}

func (r *Router) Post(path string, handler AFastHandler, middlewares ...AFastMiddleware) {
	r.method("POST", path, handler, middlewares...)
}

func (r *Router) Put(path string, handler AFastHandler, middlewares ...AFastMiddleware) {
	r.method("PUT", path, handler, middlewares...)
}

func (r *Router) Patch(path string, handler AFastHandler, middlewares ...AFastMiddleware) {
	r.method("PATCH", path, handler, middlewares...)
}

func (r *Router) Delete(path string, handler AFastHandler, middlewares ...AFastMiddleware) {
	r.method("DELETE", path, handler, middlewares...)
}

func (r *Router) Group(path string, middlewares ...AFastMiddleware) *Router {
	return r
}

func (r *Router) Use(middlewares ...AFastMiddleware) {
}

func (r *Router) View(path string, v AFastView, end bool, middlewares ...AFastMiddleware) {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "" {
			parts = append(parts[:i], parts[i+1:]...)
		}
	}

	if len(parts) == 0 {
		r.view = &view{
			view:        v,
			middlewares: middlewares,
		}
		if !end {
			if r.children == nil {
				r.children = &map[string]*Router{}
			}
			if _, ok := (*r.children)[":primary:number"]; !ok {
				(*r.children)[":primary:number"] = NewRouter()
				(*r.children)[":primary:number"].View(path, v, true, middlewares...)
			}
		}
	} else {
		first := parts[0]
		rest := parts[1:]
		if r.children == nil {
			r.children = &map[string]*Router{}
		}
		if _, ok := (*r.children)[first]; !ok {
			(*r.children)[first] = NewRouter()
		}
		(*r.children)[first].View(strings.Join(rest, "/"), v, end, middlewares...)
	}
}

func (r *Router) Websocket(path string, w AFastWebsocket, middlewares ...AFastMiddleware) {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "" {
			parts = append(parts[:i], parts[i+1:]...)
		}
	}

	if len(parts) == 0 {
		r.websocket = &ws{
			websocket:   w,
			middlewares: middlewares,
		}
	} else {
		first := parts[0]
		rest := parts[1:]
		if r.children == nil {
			r.children = &map[string]*Router{}
		}
		if _, ok := (*r.children)[first]; !ok {
			(*r.children)[first] = NewRouter()
		}
		(*r.children)[first].Websocket(strings.Join(rest, "/"), w, middlewares...)
	}
}

func (r *Router) method(method string, path string, handler AFastHandler, middlewares ...AFastMiddleware) {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "" {
			parts = append(parts[:i], parts[i+1:]...)
		}
	}

	if len(parts) == 0 {
		mergeMiddlewares := append(r.middlewares, middlewares...)
		(*r.router)[method] = handle{
			handle:      handler,
			middlewares: mergeMiddlewares,
		}
	} else {
		first := parts[0]
		rest := parts[1:]
		if r.children == nil {
			r.children = &map[string]*Router{}
		}
		if _, ok := (*r.children)[first]; !ok {
			(*r.children)[first] = NewRouter()
		}
		(*r.children)[first].method(method, strings.Join(rest, "/"), handler, middlewares...)
	}
}

func (r *Router) Index(path string, params *map[string]interface{}) (*Router, *map[string]interface{}, error) {
	if params == nil {
		params = &map[string]interface{}{}
	}
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "" {
			parts = append(parts[:i], parts[i+1:]...)
		}
	}

	if len(parts) == 0 {
		return r, params, nil
	} else {
		first := parts[0]
		rest := parts[1:]

		flat := false
		if r.children == nil {
			flat = true
		}
		if _, ok := (*r.children)[first]; !ok {
			flat = true
		}

		if flat {
			key := ""
			for k := range *r.children {
				if strings.HasPrefix(k, ":") {
					key = k
					break
				}
			}
			fmt.Println("flat", key)
			if key == "" {
				return nil, nil, nil
			}
			part := strings.Split(key, ":")
			if len(part) == 3 {
				switch part[2] {
				case "number":
					value, err := strconv.Atoi(path)
					if err != nil {
						return nil, nil, err
					}
					(*params)[part[1]] = value
				default:
					(*params)[part[1]] = first
				}
			}
			return (*r.children)[key].Index(strings.Join(rest, "/"), params)
		}

		return (*r.children)[first].Index(strings.Join(rest, "/"), params)
	}
}
