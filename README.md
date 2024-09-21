# AFast Web Framework


# Example

```go
package main

import (
	"fmt"

	"github.com/ahriroot/afast-go"
)

type CustomeMiddleware struct {
}

func (m CustomeMiddleware) Request(req *afast.Request) (*afast.Request, *afast.Response, error) {
	fmt.Printf("Before middleware: %v\n", req)
	return nil, nil, nil
}

func (m CustomeMiddleware) Response(req *afast.Request, res *afast.Response) (*afast.Response, error) {
	fmt.Printf("After middleware: %v\n", req)
	return nil, nil
}

type CustomeView struct {
}

func (v CustomeView) Get(req *afast.Request) (*afast.Response, error) {
	fmt.Printf("Received request2: %v\n", req)
	return &afast.Response{}, nil
}

func Handle(req *afast.Request) (*afast.Response, error) {
	fmt.Printf("Received request: %v\n", req)
	return &afast.Response{}, nil
}

type CustomeWebsocket struct {
}

func (ws CustomeWebsocket) OnError(req *afast.Request, err error) {
	fmt.Printf("Received error: %v\n", err)
}

func (ws CustomeWebsocket) OnConnect(req *afast.Request, sender *afast.AFastWebsocketSender) error {
	fmt.Printf("Received connect: %v\n", req)
	return nil
}

func (ws CustomeWebsocket) OnMessage(req *afast.Request, message string, sender *afast.AFastWebsocketSender) (*string, error) {
	fmt.Printf("Received message: %v\n", message)
	sender.Send("Hello, world!")
	sender.SendJson(map[string]string{"key": "value"})
	return nil, nil
}

func main() {
	app := afast.NewApp()

	app.Get("/a/b/c", Handle, CustomeMiddleware{})
	app.View("/a/b/d", CustomeView{}, false)
	app.Websocket("/ws", CustomeWebsocket{})

	m := app.Map()
	fmt.Println(m)

	app.Listen("localhost:8080")
}

```