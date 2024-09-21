package afast

import (
	"golang.org/x/net/websocket"
)

type AFastMiddleware interface {
	Request(req *Request) (*Request, *Response, error)
	Response(req *Request, resp *Response) (*Response, error)
}

type Address = string
type AFastHandler = func(req *Request) (*Response, error)
type AFastWebsocketSender struct {
	ws *websocket.Conn
}

func (s *AFastWebsocketSender) Send(message string) error {
	return websocket.Message.Send(s.ws, message)
}

func (s *AFastWebsocketSender) SendJson(data interface{}) error {
	return websocket.JSON.Send(s.ws, data)
}

func (s *AFastWebsocketSender) Close() error {
	return s.ws.Close()
}

type AFastWebsocket interface {
	OnError(req *Request, err error)
	OnConnect(req *Request, sender *AFastWebsocketSender) error
	OnMessage(req *Request, message string, sender *AFastWebsocketSender) (*string, error)
}
type AFastView interface {
}

type Options struct {
	Address Address
}

type App struct {
	server *Server
}

func NewApp() *App {
	return &App{
		server: &Server{
			router: NewRouter(),
		},
	}
}

func (a *App) Run(options Options) {
	a.server.Run(options)
}

func (a *App) Listen(addr Address) {
	a.server.Run(Options{
		Address: addr,
	})
}

func (a *App) Map() string {
	return a.server.router.Map()
}

func (a *App) MapJson() string {
	return a.server.router.MapJson()
}

func (a *App) Get(path string, handler AFastHandler, middlewares ...AFastMiddleware) {
	a.server.router.Get(path, handler, middlewares...)
}

func (a *App) Post(path string, handler AFastHandler, middlewares ...AFastMiddleware) {
	a.server.router.Post(path, handler, middlewares...)
}

func (a *App) Put(path string, handler AFastHandler, middlewares ...AFastMiddleware) {
	a.server.router.Put(path, handler, middlewares...)
}

func (a *App) Patch(path string, handler AFastHandler, middlewares ...AFastMiddleware) {
	a.server.router.Patch(path, handler, middlewares...)
}

func (a *App) Delete(path string, handler AFastHandler, middlewares ...AFastMiddleware) {
	a.server.router.Delete(path, handler, middlewares...)
}

func (a *App) Group(path string, middlewares ...AFastMiddleware) {
	a.server.router.Group(path, middlewares...)
}

func (a *App) Use(middlewares ...AFastMiddleware) {
	a.server.router.Use(middlewares...)
}

func (a *App) View(path string, v AFastView, end bool, middlewares ...AFastMiddleware) {
	a.server.router.View(path, v, end, middlewares...)
}

func (a *App) Websocket(path string, websocket AFastWebsocket, middlewares ...AFastMiddleware) {
	a.server.router.Websocket(path, websocket, middlewares...)
}
