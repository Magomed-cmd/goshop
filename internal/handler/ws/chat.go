package ws

import (
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// CheckOrigin позволяет принимать соединения от любого origin
	// В продакшене следует ограничить разрешенные origins
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
