package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// address => ws_connection -> last_active = we can clean up inactive connections after some time

// TODO
// 1. maybe we can add the user address as a path param on notifications endpoint
// 2. maybe we add a table of notifications that we can sen to the client besides websockets

type WSConn struct {
	*websocket.Conn
	LastActive time.Time
}

// var connections = make(map[string]*websocket.Conn)
var connections = make(map[string]*WSConn)

type Message struct {
	Address string
	Msg     []byte
}

var messages []Message

func main() {

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			// Replace with your origin validation logic
			return true // Allow all origins for demonstration
		},
	}

	g := gin.Default()

	g.Static("/public", "./public")

	g.POST("/send", func(c *gin.Context) {
		var payload struct {
			Address string `json:"address"`
		}
		c.BindJSON(&payload)
		message := Message{payload.Address, []byte(fmt.Sprintf("Transaction done for: %s", payload.Address))}
		messages = append(messages, message)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	g.GET("/notifications", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			fmt.Println("Upgrade:", err)
			return
		}

		fmt.Println("Client connected:", conn.RemoteAddr())

		// Read address and store connection
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Read:", err)
		}

		connections[string(message)] = &WSConn{conn, time.Now()}
	})

	// var previousNumGoroutine int
	// go func() {
	// 	for {
	// 		<-time.After(1 * time.Second)
	// 		currentNumGoroutine := runtime.NumGoroutine()
	// 		if currentNumGoroutine != previousNumGoroutine {
	// 			fmt.Println("NumGoroutine: ", currentNumGoroutine)
	// 			previousNumGoroutine = currentNumGoroutine
	// 		}
	// 	}
	// }()

	// consume messages and send to client
	go func() {
		for {
			<-time.After(1 * time.Second)
			if len(messages) > 0 {
				msg := messages[0]
				messages = messages[1:]
				send(msg.Address, msg.Msg)
			}
		}
	}()

	// remove inactive connections
	go func() {
		for {
			for address, conn := range connections {
				if time.Since(conn.LastActive) > 15*time.Second {
					fmt.Println("Client disconnected:", conn.RemoteAddr())
					conn.Close()
					delete(connections, address)
				}
			}
		}
	}()

	g.Run(":3000")
}

func send(address string, msg []byte) {
	conn, ok := connections[address]
	if ok {
		err := conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			fmt.Println("Write:", err)
			delete(connections, address)
		}
		conn.LastActive = time.Now()
	}
}
