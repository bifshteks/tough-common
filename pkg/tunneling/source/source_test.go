package source

import (
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"testing"
)

func TestSourceStructsImplementInterfaces(t *testing.T) {
	var _ Source = NewWSConn(nil, websocket.TextMessage, logrus.StandardLogger())
	var _ Source = NewTCPConnection(nil, logrus.StandardLogger())

	ws := NewWS("", websocket.TextMessage, nil, logrus.StandardLogger())
	var _ Source = ws
	var _ NetworkSource = ws

	tcp := NewTCP("", logrus.StandardLogger())
	var _ Source = tcp
	var _ NetworkSource = tcp
}
