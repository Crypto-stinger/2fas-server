package common

import (
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/twofas/2fas-server/internal/common/logging"
	"github.com/twofas/2fas-server/internal/common/recovery"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4 * 1024,
	WriteBufferSize: 4 * 1024,
	CheckOrigin: func(r *http.Request) bool {
		allowedOrigin := os.Getenv("WEBSOCKET_ALLOWED_ORIGIN")

		if allowedOrigin != "" {
			return r.Header.Get("Origin") == allowedOrigin
		}

		return true
	},
}

type ConnectionHandler struct {
	hubs *hubPool
	mtx  *sync.Mutex
}

func NewConnectionHandler() *ConnectionHandler {
	return &ConnectionHandler{
		hubs: newHubPool(),
		mtx:  &sync.Mutex{},
	}
}

func (h *ConnectionHandler) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		channel := c.Request.URL.Path

		log := logging.FromContext(c.Request.Context()).WithField("channel", channel)

		log.Info("New channel subscriber")

		h.serveWs(c.Writer, c.Request, channel, log)
	}
}

func (h *ConnectionHandler) serveWs(w http.ResponseWriter, r *http.Request, channel string, log logging.FieldLogger) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("Failed to upgrade connection: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	client, _ := h.hubs.registerClient(channel, conn)

	go recovery.DoNotPanic(func() {
		client.writePump()
	})

	go recovery.DoNotPanic(func() {
		client.readPump(log)
	})

	go recovery.DoNotPanic(func() {
		disconnectAfter := 3 * time.Minute
		timeout := time.After(disconnectAfter)

		<-timeout
		log.Info("Connection closed after", disconnectAfter)

		client.hub.unregisterClient(client)
		client.conn.Close()
	})
}
