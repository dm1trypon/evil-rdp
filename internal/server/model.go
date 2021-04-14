package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type Server struct {
	lc          string
	port        int
	wsCodes     map[string]int
	server      *http.Server
	router      *mux.Router
	upgrader    websocket.Upgrader
	clients     map[string]*wsConnData
	msgs        chan []byte
	newClient   chan string
	serverError *chan error
}

type wsConnData struct {
	wsStream      *websocket.Conn
	wsInteractive *websocket.Conn
}
