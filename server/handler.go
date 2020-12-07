package server

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
)

type websocketHandler struct {
	upgrader *websocket.Upgrader
	Conn     *Conn

}

// RegisterMessage defines message struct client send after connect
// to the server.
type RegisterMessage struct {
	BrowserMessage string
}

// First try to upgrade connection to websocket. If success, connection will
// be kept until client send close message or server drop them.
func (wh *websocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wsConn, err := wh.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer wsConn.Close()

	// handle Websocket request
	conn := NewConn(wsConn)
	conn.AfterReadFunc = func(messageType int, r io.Reader) {
		var rm RegisterMessage
		decoder := json.NewDecoder(r)
		if err := decoder.Decode(&rm); err != nil {
			return
		}
		wh.Conn = conn

		// for control
		switch rm.BrowserMessage {
		case "register":
			fmt.Println("Connection established.")
			break
		case "start":
			fmt.Println("Start to simulate.")
			producer := GetInstance()
			producer.SetConn(conn)
			producer.Start()
			break
		case "pause":
			fmt.Println("Pause.")
			producer := GetInstance()
			producer.Pause()
			break


		default:
			break
		}

	}

	conn.BeforeCloseFunc = func() {
		// unbind
		wh.Conn = nil
	}

	conn.Listen()
}



