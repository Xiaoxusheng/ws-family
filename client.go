package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
	"ws/ping"
)

var Dialer = &websocket.Dialer{
	NetDial:           nil,
	NetDialContext:    nil,
	NetDialTLSContext: nil,
	Proxy:             nil,
	TLSClientConfig:   nil,
	HandshakeTimeout:  0,
	ReadBufferSize:    0,
	WriteBufferSize:   0,
	WriteBufferPool:   nil,
	Subprotocols:      nil,
	EnableCompression: false,
	Jar:               nil,
}

type Message struct {
	MessageType int    //消息类型
	Data        []byte //数据
}

type ClientWs struct {
	Ws    *websocket.Conn
	Read  chan *Message
	Write chan *Message
}

func CreateClient() (*websocket.Conn, *http.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // 确保在函数结束时取消上下文

	// 创建 WebSocket 连接
	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second, // 设置握手超时时间
	}

	// 发送请求头（如果需要，可以在这里添加自定义头）
	headers := http.Header{}
	// headers.Add("Authorization", "Bearer your_token") // 添加自定义头（示例）

	c, h, err := dialer.DialContext(ctx, "ws://xlei.fun:8080/ws", headers)
	if err != nil {
		log.Fatalf("Dial error: %v", err) // 使用 Fatalf 方便调试，打印错误后退出
		return nil, nil, err
	}
	log.Println("Dial success client addr:", h.Request.URL.String())
	return c, h, nil
}

func NewClientWs() *ClientWs {
	clientWs := new(ClientWs)
	//获取链接
	ws, _, err := CreateClient()
	if err != nil {
		panic(err)
	}
	clientWs.Ws = ws
	clientWs.Read = make(chan *Message, 100)
	clientWs.Write = make(chan *Message, 100)
	return clientWs
}

func (c *ClientWs) read() {
	for {
		messageType, message, err := c.Ws.ReadMessage()
		if err != nil {
			err = c.Ws.Close()
			if err != nil {
				log.Println("close con fail err:", err)
				return
			}
			return
		}
		c.Read <- &Message{
			MessageType: messageType,
			Data:        message,
		}
	}
}

func (c *ClientWs) write() {
	for {
		select {
		case data := <-c.Write:
			err := c.Ws.WriteMessage(data.MessageType, data.Data)
			if err != nil {
				err = c.Ws.Close()
				if err != nil {
					log.Println("close con fail err:", err)
					return
				}
				continue
			}
		}
	}
}

func (c *ClientWs) play() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	go c.read()
	go c.write()
	go c.Ws.SetPingHandler(func(appData string) error {
		//log.Printf("Received ping with data: %s", appData)
		ping.Pong(c.Ws)
		// 可以在这里做一些处理，比如记录日志
		return nil // 返回 nil 表示处理成功
	})

	for {
		select {
		case message := <-c.Read:
			log.Println(message.MessageType, string(message.Data))
			//c.Write <- message
		case <-interrupt:
			log.Println("interrupt")
			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.Ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			os.Exit(1)
		}
	}
}

func scan(c *ClientWs) {
	for {
		fmt.Print("请输入聊天内容：")
		reader := bufio.NewReader(os.Stdin)
		messageData, err := reader.ReadString('\n')
		if err != nil {
			log.Println("read:", err)
			continue
		}
		trimmedMessage := strings.TrimSpace(messageData)
		if len(trimmedMessage) == 0 {
			continue
		}
		c.Write <- &Message{
			MessageType: websocket.TextMessage,
			Data:        []byte(trimmedMessage),
		}
	}
}

//func client() {
//	c, h, err := CreateClient()
//	if err != nil {
//		panic(err)
//	}
//	defer c.Close()
//	interrupt := make(chan os.Signal, 1)
//	signal.Notify(interrupt, os.Interrupt)
//	// 打印连接信息
//	fmt.Printf("Connected to %s\n", h.Request.URL.String())
//	go c.SetPingHandler(func(appData string) error {
//		log.Printf("Received ping with data: %s", appData)
//		ping.Pong(c)
//
//		// 可以在这里做一些处理，比如记录日志
//		return nil // 返回 nil 表示处理成功
//	})
//
//	go func() {
//		for {
//			select {
//			case <-interrupt:
//				log.Println("interrupt")
//
//				// Cleanly close the connection by sending a close message and then
//				// waiting (with timeout) for the server to close the connection.
//				err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
//				if err != nil {
//					log.Println("write close:", err)
//					return
//				}
//				os.Exit(1)
//			}
//
//		}
//	}()
//
//	go func(c *websocket.Conn) {
//		for {
//			// 示例：读取消息
//			_, msg, err := c.ReadMessage()
//			if err != nil {
//				log.Printf("Read error: %v", err)
//				return
//			}
//			fmt.Printf("Received: %s\n", msg)
//		}
//	}(c)
//	// 这里可以添加你想要的处理逻辑，比如读取和写入消息
//	for {
//		time.Sleep(time.Second * 4)
//		// 示例：发送消息
//		if err := c.WriteMessage(websocket.TextMessage, []byte("Hello from client!")); err != nil {
//			log.Printf("Write error: %v", err)
//			continue
//		}
//
//	}
//}

func main() {
	c := NewClientWs()
	defer c.Ws.Close()
	c.play()
}
