package ping

import (
	"github.com/gorilla/websocket"
	"log"
	"time"
	"ws/config"
)

// Ping  程序
func Ping(con *websocket.Conn, dur time.Duration,cs  *config.WebSocketConnect) {
	t := time.NewTicker(time.Second * dur)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			err := con.WriteMessage(websocket.PingMessage, []byte("ping"))
			if err != nil {
				log.Println(err)
				cs.Mux.Lock()
				delete(cs.Con,con)
				cs.Mux.Unlock()
				return
			}
		}
	}
}

func Pong(con *websocket.Conn) {
	err := con.WriteMessage(websocket.PongMessage, []byte("pong"))
	if err != nil {
		log.Println(err)
		return
	}
}
