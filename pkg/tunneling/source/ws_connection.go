package source

import (
	"context"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"time"
)

type CWS struct { // connection webSocket
	conn    *websocket.Conn
	reader  chan []byte
	msgType int
}

func NewCWS(conn *websocket.Conn, msgType int) *CWS {
	return &CWS{
		conn:    conn,
		msgType: msgType,
		reader:  make(chan []byte),
	}
}

func (ws *CWS) GetName() string {
	return "CWS"
}

func (ws *CWS) GetReader() chan []byte {
	return ws.reader
}
func (ws *CWS) Connect(ctx context.Context, cancelFunc context.CancelFunc) {
	// don't need to implement this, because we already
	// initialized this structure with active connection.
}

func (ws *CWS) Start(ctx context.Context, cancel context.CancelFunc) {
	defer logrus.Debugln("gorounting wsConn.Start() ends")

	// cannot set readDeadLine - https://github.com/gorilla/websocket/issues/474,
	// so use this goroutine
	go func() {
		<-ctx.Done()
		ws.Close()
	}()
	for {
		_, message, err := ws.conn.ReadMessage()
		if err == nil {
			ws.reader <- message
			continue
		}
		logrus.Errorln("Cannot read from ws conn", err)
		cancel()
		return
	}
}

func (ws *CWS) Close() {
	defer logrus.Debugln("wsConn.Close() ends")
	if ws.conn == nil {
		// if we closed session before even got the connection
		return
	}
	_ = ws.conn.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))

	// waiting (with timeout) for the server to close the connection.
	<-time.After(time.Second)
	_ = ws.conn.Close()
	close(ws.reader)
}

func (ws *CWS) Write(msg []byte) {
	err := ws.conn.WriteMessage(ws.msgType, msg)
	if err != nil {
		logrus.Errorln("Could not write to ws connection", err, msg)
	}
}
