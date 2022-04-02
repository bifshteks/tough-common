package source

import (
	"github.com/gorilla/websocket"
	"testing"
)

func TestSourceStructsImplementInterfaces(t *testing.T) {
	var _ Source = NewWSConn(nil, websocket.TextMessage)
	var _ Source = NewTCPConnection(nil)

	ws := NewWS("", websocket.TextMessage, nil)
	var _ Source = ws
	var _ NetworkSource = ws

	tcp := NewTCP("")
	var _ Source = tcp
	var _ NetworkSource = tcp
}
