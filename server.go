package afast

import (
	"net/http"
	"reflect"
	"strings"

	"golang.org/x/net/websocket"
)

type Server struct {
	router *Router
}

func (s *Server) Run(options Options) error {
	http.Handle("/", &Handle{
		router: s.router,
	})

	return http.ListenAndServe(options.Address, nil)
}

type Handle struct {
	router *Router
}

func (h *Handle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	url := r.URL.Path

	// 匹配路由
	ro, p, e := h.router.Index(url, nil)
	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(e.Error()))
		return
	}

	if ro == nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 Not Found"))
		return
	}

	// 当前路径未设置处理函数，且未设置试图，且未设置websocket
	if len(*ro.router) == 0 && ro.view == nil && ro.websocket == nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 Not Found"))
		return
	}

	// 构造请求对象
	req := NewRequest(r, p)

	if r.Header.Get("Upgrade") == "websocket" {
		if ro.websocket == nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("404 Not Found"))
			return
		}

		// 响应websocket请求
		ws := websocket.Handler(func(ws *websocket.Conn) {
			defer ws.Close()

			sender := AFastWebsocketSender{
				ws: ws,
			}

			for {
				var msg string
				if err := websocket.Message.Receive(ws, &msg); err != nil {
					ro.websocket.websocket.OnError(req, err)
					break
				}

				resp, err := ro.websocket.websocket.OnMessage(req, msg, &sender)
				if err != nil {
					break
				}

				if resp != nil {
					if err := websocket.Message.Send(ws, resp); err != nil {
						break
					}
				}
			}
		})
		ws.ServeHTTP(w, r)
		return
	}

	// 处理请求
	if handle, ok := (*ro.router)[method]; ok {
		// 请求中间件
		for i := 0; i < len(handle.middlewares); i++ {
			mw := handle.middlewares[i]
			mwreq, mwresp, mwerr := mw.Request(req)
			if mwerr != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(mwerr.Error()))
				return
			} else if mwresp != nil {
				resp := mwresp.ToBytes()
				w.WriteHeader(mwresp.Status)
				w.Write(resp)
				return
			} else {
				req = mwreq
			}
		}

		// 调用处理函数
		resp, err := handle.handle(req)

		// 响应中间件
		for i := len(handle.middlewares) - 1; i >= 0; i-- {
			mw := handle.middlewares[i]
			mwresp, mwerr := mw.Response(req, resp)
			if mwerr != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(mwerr.Error()))
				return
			}
			resp = mwresp
		}

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		if resp == nil {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(""))
		} else {
			if resp.Status == 0 {
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(resp.Status)
			}
			w.Write(resp.ToBytes())
		}
		return
	} else {
		if ro.view == nil {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("405 Method Not Allowed"))
			return
		}

		// 调用试图函数
		m := method[0:1] + strings.ToLower(method[1:])
		fn := reflect.ValueOf(ro.view.view).MethodByName(m)
		if !fn.IsValid() {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("405 Method Not Allowed"))
			return
		}

		result := fn.Call([]reflect.Value{reflect.ValueOf(req)})

		if len(result) == 2 {
			// 判断 reflect.Value 是不是 nil
			if !result[1].IsNil() {
				err := result[1].Interface().(error)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			if !result[0].IsNil() {
				resp := result[0].Interface().(*Response)
				if resp.Status == 0 {
					w.WriteHeader(http.StatusOK)
				} else {
					w.WriteHeader(resp.Status)
				}
				w.Write(resp.ToBytes())
				return
			} else {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(""))
				return
			}
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 Internal Server Error"))
		return
	}

	// if r.Header.Get("Upgrade") == "websocket" {
	// 	// 响应websocket请求
	// 	ws := websocket.Handler(func(ws *websocket.Conn) {
	// 		defer ws.Close()

	// 		for {
	// 			var msg string
	// 			// 读取消息
	// 			if err := websocket.Message.Receive(ws, &msg); err != nil {
	// 				break
	// 			}
	// 			// 处理消息，例如打印到控制台
	// 			println(msg)

	// 			// 回送消息
	// 			if err := websocket.Message.Send(ws, msg); err != nil {
	// 				break
	// 			}
	// 		}
	// 	})
	// 	ws.ServeHTTP(w, r)
	// } else {

	// }
}
