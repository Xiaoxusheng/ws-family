package config

import (
	"github.com/gorilla/websocket"
	"sync"
)

// WebSocketConnect 全局记录客户端连接信息
type WebSocketConnect struct {
	Mux *sync.RWMutex                //读写锁
	Con map[*websocket.Conn]struct{} //套接字链接
}

// WebSocketConfig 记录每个连接的信息
type WebSocketConfig struct {
	Ws    *websocket.Conn
	Read  chan []byte //写管道
	Write chan []byte //读管道
}
