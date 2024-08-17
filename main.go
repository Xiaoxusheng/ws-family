package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net"
	"net/http"
	"sync"
	"ws/config"
	"ws/ping"
)

var upgrader = websocket.Upgrader{
	HandshakeTimeout:  0,
	ReadBufferSize:    1024,
	WriteBufferSize:   1024,
	WriteBufferPool:   new(sync.Pool),
	Subprotocols:      nil,
	Error:             nil,
	CheckOrigin:       nil,
	EnableCompression: false,
}

var Connections = config.WebSocketConnect{
	Mux: &sync.RWMutex{},
	Con: make(map[*websocket.Conn]struct{}),
}

func handler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()
	//写进
	Connections.Mux.Lock()
	Connections.Con[conn] = struct{}{}
	Connections.Mux.Unlock()

	go ping.Ping(conn, 5, &Connections)
	conn.SetCloseHandler(func(code int, text string) error {
		fmt.Println("关闭", code, text)
		Connections.Mux.Lock()
		delete(Connections.Con, conn)
		Connections.Mux.Unlock()
		return nil
	})

	conn.SetPongHandler(func(appData string) error {
		log.Printf("Received pong with data: %s", appData)
		log.Println(Connections.Con)
		// 可以在这里做一些处理，比如记录日志
		return nil // 返回 nil 表示处理成功
	})

	for {
		message, i, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			fmt.Println(err)
			return
		}
		for c := range Connections.Con {
			err = c.WriteMessage(websocket.TextMessage, i)
			if err != nil {
				log.Println(err)
				return
			}
			log.Println(ip, message, "\n", string(i))
		}

	}
}

func main() {
	http.HandleFunc("/ws", handler)
	log.Println("websocket server running on :8080")

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
		return
	}
}
