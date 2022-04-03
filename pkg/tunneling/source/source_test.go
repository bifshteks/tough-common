package source

import (
	"github.com/bifshteks/tough_common/pkg/logutil"
	"github.com/gorilla/websocket"
	"testing"
)

func TestSourceStructsImplementInterfaces(t *testing.T) {
	var _ Source = NewWSConn(nil, websocket.TextMessage, logutil.DummyLogger)
	var _ Source = NewTCPConnection(nil, logutil.DummyLogger)

	ws := NewWS("", websocket.TextMessage, nil, logutil.DummyLogger)
	var _ Source = ws
	var _ NetworkSource = ws

	tcp := NewTCP("", logutil.DummyLogger)
	var _ Source = tcp
	var _ NetworkSource = tcp
}
