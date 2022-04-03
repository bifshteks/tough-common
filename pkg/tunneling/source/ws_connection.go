package source

import (
	"context"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"time"
)

type WSConn struct { // connection webSocket
	conn    *websocket.Conn
	reader  chan []byte
	msgType int
	logger  *logrus.Logger
}

func NewWSConn(conn *websocket.Conn, msgType int, logger *logrus.Logger) *WSConn {
	return &WSConn{
		conn:    conn,
		msgType: msgType,
		reader:  make(chan []byte),
		logger:  logger,
	}
}

func (ws *WSConn) GetReader() chan []byte {
	return ws.reader
}

func (ws *WSConn) Consume(ctx context.Context) (err error) {
	defer ws.logger.Debugln("gorounting wsConn.Start() ends")

	// cannot set readDeadLine - https://github.com/gorilla/websocket/issues/474,
	// so use this goroutine
	go func() {
		<-ctx.Done()
		ws.Close()
	}()
	for {
		_, message, err := ws.conn.ReadMessage()
		if err != nil {
			normalClosure := websocket.IsCloseError(err, websocket.CloseNormalClosure)
			if normalClosure {
				return nil
			}
			return err
		}
		ws.reader <- message
	}
}

func (ws *WSConn) Write(msg []byte) (err error) {
	return ws.conn.WriteMessage(ws.msgType, msg)
}

func (ws *WSConn) Close() {
	defer ws.logger.Debugln("wsConn.Close() ends")
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
	ws.conn = nil
	close(ws.reader)
}
