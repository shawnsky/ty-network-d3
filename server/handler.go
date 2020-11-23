package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/shawnsky/ty-network-d3/model"
	"io"
	"net/http"
	"strings"
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

		fmt.Println("Get reg")
		// discard browser msg, just bind
		wh.Conn = conn
	}

	conn.BeforeCloseFunc = func() {
		// unbind
		wh.Conn = nil
	}

	conn.Listen()
}


// ErrRequestIllegal describes error when data of the request is unaccepted.
var ErrRequestIllegal = errors.New("request data illegal")

// produceHandler defines to handle produce message request.
type produceHandler struct {
	websocketHandler *websocketHandler
}

func (s *produceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}


	// read model state
	var pm model.PushMessage
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&pm); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(ErrRequestIllegal.Error()))
		return
	}

	fmt.Println("WS server read from produce",pm)
	// continue send to ws client
	jsonData,_ := json.Marshal(pm)
	_, err := s.websocketHandler.Conn.Write(jsonData)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	result := strings.NewReader(fmt.Sprintf("message sent to client\n"))
	io.Copy(w, result)
}




