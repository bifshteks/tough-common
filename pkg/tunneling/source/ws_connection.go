package source

import (
	"context"
	"errors"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"time"
)

type WSConn struct { // connection webSocket
	conn    *websocket.Conn
	reader  chan []byte
	msgType int
}

func NewWSConn(conn *websocket.Conn, msgType int) *WSConn {
	return &WSConn{
		conn:    conn,
		msgType: msgType,
		reader:  make(chan []byte),
	}
}

func (ws *WSConn) GetReader() chan []byte {
	return ws.reader
}

func (ws *WSConn) Consume(ctx context.Context) (err error) {
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
		// TODO check if error is really an error, not just end of connection
		return errors.New("Cannot read from ws conn: " + err.Error())
	}
}

func (ws *WSConn) Close() {
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

func (ws *WSConn) Write(msg []byte) (err error) {
	return ws.conn.WriteMessage(ws.msgType, msg)
}
