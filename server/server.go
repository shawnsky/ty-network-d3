package server

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"strings"
)

const (
	serverDefaultWSPath   = "/ws"
	serverDefaultPushPath = "/produce"
)

var defaultUpgrader = &websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(*http.Request) bool {
		return true
	},
}

// Server defines parameters for running websocket server.
type Server struct {
	// Address for server to listen on
	Addr string

	// Path for websocket request, default "/ws".
	WSPath string

	// Path for produce message, default "/produce".
	PushPath string

	// Upgrader is for upgrade connection to websocket connection using
	// "github.com/gorilla/websocket".
	//
	// If Upgrader is nil, default upgrader will be used. Default upgrader is
	// set ReadBufferSize and WriteBufferSize to 1024, and CheckOrigin always
	// returns true.
	Upgrader *websocket.Upgrader

	wh *websocketHandler
	ph *produceHandler
}

// ListenAndServe listens on the TCP network address and handle websocket
// request.
func (s *Server) ListenAndServe() error {

	// websocket request handler
	wh := websocketHandler{
		upgrader: defaultUpgrader,
	}
	if s.Upgrader != nil {
		wh.upgrader = s.Upgrader
	}

	s.wh = &wh
	http.Handle(s.WSPath, s.wh)

	// produce request handler
	ph := produceHandler{
		websocketHandler: &wh,
	}

	s.ph = &ph
	http.Handle(s.PushPath, s.ph)

	return http.ListenAndServe(s.Addr, nil)
}


// Check parameters of Server, returns error if fail.
func (s Server) check() error {
	if !checkPath(s.WSPath) {
		return fmt.Errorf("WSPath: %s not illegal", s.WSPath)
	}
	if !checkPath(s.PushPath) {
		return fmt.Errorf("PushPath: %s not illegal", s.PushPath)
	}
	if s.WSPath == s.PushPath {
		return errors.New("WSPath is equal to PushPath")
	}

	return nil
}

// NewServer creates a new Server.
func NewServer(addr string) *Server {
	return &Server{
		Addr:     addr,
		WSPath:   serverDefaultWSPath,
		PushPath: serverDefaultPushPath,
	}
}

func checkPath(path string) bool {
	if path != "" && !strings.HasPrefix(path, "/") {
		return false
	}
	return true
}

