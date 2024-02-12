package main

import (
	"net"
	"net/http"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/google/uuid"
)

type User struct {
	ID   string
	Name string
	Conn *net.Conn
}

type Message struct {
	From    string
	To      string
	Content string
}

func main() {
	users := make(map[string]User)

	http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// upgrade connection
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			// handle error
		}

		// track connection
		id := uuid.New().String()
		users[id] = User{
			ID:   id,
			Conn: &conn,
		}

		// handle requests
		// warn: new goroutine with every request?
		go func() {
			// close connection on shutdown
			defer conn.Close()

			// ...
			for {
				msg, op, err := wsutil.ReadClientData(conn)
				if err != nil {
					// handle error
				}

				// ...
				for _, user := range users {
					// if id == user.ID {
					// 	continue
					// }
					// log.Printf("send to %s", id)
					// msg = []byte(fmt.Sprintf("%s: %s", id, string(msg)))
					err := wsutil.WriteServerMessage(*user.Conn, op, msg)
					if err != nil {
						// handle error
					}
				}

				// err = wsutil.WriteServerMessage(conn, op, msg)
				// if err != nil {
				// 	// handle error
				// }
			}
		}()
	}))
}

func UnmarshalMessage(msg []byte) (Message, error) {
	return Message{}, nil
}
