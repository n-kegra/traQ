package ws

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
)

const (
	writeWait          = 5 * time.Second
	pongWait           = 60 * time.Second
	pingPeriod         = (pongWait * 9) / 10
	maxReadMessageSize = 1 << 9 // 512B
	messageBufferSize  = 256
)

var (
	json     = jsoniter.ConfigFastest
	upgrader = &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
)
