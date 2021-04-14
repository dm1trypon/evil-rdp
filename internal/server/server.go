package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	logger "github.com/dm1trypon/easy-logger"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

func (s *Server) Create(serverError *chan error, port int) *Server {
	s = &Server{
		lc:          "SERVER",
		port:        port,
		wsCodes:     map[string]int{},
		server:      &http.Server{},
		router:      &mux.Router{},
		upgrader:    websocket.Upgrader{},
		clients:     map[string]*wsConnData{},
		msgs:        make(chan []byte),
		newClient:   make(chan string),
		serverError: serverError,
	}

	s.setWSCodes()
	s.setUpgrader()
	s.setRouter()
	go s.startServer()

	return s
}

func (s *Server) GetMsgsChan() *chan []byte {
	return &s.msgs
}

func (s *Server) GetNewClient() *chan string {
	return &s.newClient
}

func (s *Server) GetNumConnectedClients() int {
	return len(s.clients)
}

func (s *Server) startServer() {
	logger.Debug(s.lc, "WS Server setup")

	s.server = &http.Server{
		Addr:              fmt.Sprint(":", strconv.Itoa(s.port)),
		Handler:           s.router,
		TLSConfig:         nil,
		ReadTimeout:       1000 * time.Second,
		ReadHeaderTimeout: 0,
		WriteTimeout:      1000 * time.Second,
		IdleTimeout:       0,
		MaxHeaderBytes:    0,
		TLSNextProto:      map[string]func(*http.Server, *tls.Conn, http.Handler){},
		ConnState: func(net.Conn, http.ConnState) {
		},
		ErrorLog:    &log.Logger{},
		BaseContext: func(net.Listener) context.Context { return context.Background() },
		ConnContext: func(ctx context.Context, c net.Conn) context.Context { return context.Background() },
	}

	logger.Info(s.lc, fmt.Sprint("WS Server started at port ", s.port))

	if err := s.server.ListenAndServe(); err != nil {
		logger.Critical(s.lc, fmt.Sprint("Failed starting WS Server: ", err.Error()))
		*s.serverError <- err
		return
	}
}

func (s *Server) setUpgrader() {
	logger.Debug(s.lc, "Upgrader setup")
	s.upgrader = websocket.Upgrader{
		HandshakeTimeout: 0,
		ReadBufferSize:   0,
		WriteBufferSize:  0,
		WriteBufferPool:  nil,
		Subprotocols:     []string{},
		Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
		},
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		EnableCompression: false,
	}
}

func (s *Server) setWSCodes() {
	s.wsCodes = map[string]int{
		"CloseNormalClosure":           1000,
		"CloseGoingAway":               1001,
		"CloseProtocolError":           1002,
		"CloseUnsupportedData":         1003,
		"CloseNoStatusReceived":        1005,
		"CloseAbnormalClosure":         1006,
		"CloseInvalidFramePayloadData": 1007,
		"ClosePolicyViolation":         1008,
		"CloseMessageTooBig":           1009,
		"CloseMandatoryExtension":      1010,
		"CloseInternalServerErr":       1011,
		"CloseServiceRestart":          1012,
		"CloseTryAgainLater":           1013,
		"CloseTLSHandshake":            1015,
	}
}

func (s *Server) setRouter() {
	logger.Debug(s.lc, "Router setup")
	s.router = mux.NewRouter()
	s.router.HandleFunc("/stream", s.onStream)
	s.router.HandleFunc("/interactive", s.onInteractive)
}

func (s *Server) onInteractive(w http.ResponseWriter, r *http.Request) {
	loggerMask := fmt.Sprint("[INTERACTIVE][", r.RemoteAddr, "]")
	wsConn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error(s.lc, fmt.Sprint(loggerMask, " Upgrader error: ", err.Error()))
		return
	}

	defer wsConn.Close()

	key := r.URL.Query().Get("key")
	loggerMask = fmt.Sprint(loggerMask, "[", key, "]")

	if len(key) < 1 {
		logger.Error(s.lc, fmt.Sprint(loggerMask, " Empty param 'key'"))
		return
	}

	if _, ok := s.clients[key]; ok {
		logger.Error(s.lc, fmt.Sprint(loggerMask, " WS client is already exist"))
		return
	}

	wsConnData := &wsConnData{
		wsInteractive: wsConn,
		wsStream:      nil,
	}

	s.clients[key] = wsConnData

	logger.Info(s.lc, fmt.Sprint(loggerMask, " Client connected"))
	s.newClient <- wsConn.RemoteAddr().String()
	s.reader(wsConn, loggerMask)
}

func (s *Server) onStream(w http.ResponseWriter, r *http.Request) {
	loggerMask := fmt.Sprint("[STREAM][", r.RemoteAddr, "]")
	wsConn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error(s.lc, fmt.Sprint(loggerMask, " Upgrader error: ", err.Error()))
		return
	}

	defer wsConn.Close()

	key := r.URL.Query().Get("key")
	loggerMask = fmt.Sprint(loggerMask, "[", key, "]")
	if len(key) < 1 {
		logger.Error(s.lc, fmt.Sprint(loggerMask, " Empty param 'key'"))
		return
	}

	if _, ok := s.clients[key]; !ok {
		logger.Error(s.lc, fmt.Sprint(loggerMask, " WS client is not authorized"))
		return
	}

	if s.clients[key].wsStream != nil {
		logger.Error(s.lc, fmt.Sprint(loggerMask, " WS client is already connected"))
		return
	}

	logger.Info(s.lc, fmt.Sprint(loggerMask, " Client connected"))
	s.clients[key].wsStream = wsConn
	s.reader(wsConn, loggerMask)
}

func (s *Server) reader(wsConn *websocket.Conn, loggerMask string) {
	for {
		_, message, err := wsConn.ReadMessage()
		if err != nil {
			logger.Error(s.lc, fmt.Sprint(loggerMask, " WS reading message error: ", err.Error()))
			s.deleteClient(wsConn, loggerMask)
			break
		}

		if s.isWSError(err) {
			logger.Warning(s.lc, fmt.Sprint(loggerMask, " WS handle error: ", err.Error()))
			s.deleteClient(wsConn, loggerMask)
			break
		}

		s.msgs <- message
	}
}

func (s *Server) deleteClient(wsConn *websocket.Conn, loggerMask string) {
	for key, wsConnData := range s.clients {
		if wsConnData.wsInteractive == wsConn || wsConnData.wsStream == wsConn {
			logger.Error(s.lc, fmt.Sprint(loggerMask,
				" There was a client disconnection, removing all child connections."))

			delete(s.clients, key)
			return
		}
	}
}

func (s *Server) isWSError(err error) bool {
	for _, code := range s.wsCodes {
		if websocket.IsUnexpectedCloseError(err, code) {
			return true
		}
	}

	return false
}

func (s *Server) StreamSend(body []byte) {
	for key, wsConnData := range s.clients {
		if wsConnData.wsStream == nil {
			continue
		}

		if err := wsConnData.wsStream.WriteMessage(2, body); err != nil {
			logger.Error(s.lc, fmt.Sprint("[", key, "] Error sending message: ", err.Error()))
		}
	}
}

func (s *Server) InteractiveSend(body []byte, address string) {
	for key, wsConnData := range s.clients {
		if wsConnData.wsInteractive == nil {
			continue
		}

		if wsConnData.wsInteractive.RemoteAddr().String() != address {
			continue
		}

		logger.Info(s.lc, fmt.Sprint("[INTERACTIVE][", address, "] SENT: ", string(body)))

		if err := wsConnData.wsInteractive.WriteMessage(1, body); err != nil {
			logger.Error(s.lc, fmt.Sprint("[", key, "] Error sending message: ", err.Error()))
		}
	}
}
