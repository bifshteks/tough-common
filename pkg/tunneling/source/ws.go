package source

import (
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"net/http"
)

type WS struct {
	url           string
	conn          *websocket.Conn
	reader        chan []byte
	msgType       int
	requestHeader http.Header
}

func NewWS(url string, msgType int, requestHeader http.Header) *WS {
	return &WS{
		url:           url,
		conn:          nil,
		reader:        make(chan []byte),
		msgType:       msgType,
		requestHeader: requestHeader,
	}
}

func (ws *WS) GetUrl() string {
	return ws.url
}

func (ws *WS) GetReader() chan []byte {
	return ws.reader
}

func (ws *WS) Connect(ctx context.Context) (err error) {
	logrus.Debugln("ws.Connect()")
	select {
	case <-ctx.Done():
		return nil
	default:
		c, resp, err := websocket.DefaultDialer.DialContext(ctx, ws.url, ws.requestHeader)
		if resp != nil {
			logrus.Debugln("ws dial response", resp.StatusCode, err)
			if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnauthorized {
				err = errors.New(
					fmt.Sprintf(
						"ws connection cannot connect to %s, response status is %d",
						ws.url, resp.StatusCode,
					),
				)
				return NewFatalConnectError(err)
			}
		}
		if err != nil {
			logrus.Errorln("Cannot connect to ws", ws.url, err)
			return err
		}
		ws.conn = c
		logrus.Infoln("ws connected to", ws.url)
		// cannot set readDeadLine - https://github.com/gorilla/websocket/issues/474,
		// so use this goroutine
		go func() {
			<-ctx.Done()
			ws.Close()
		}()
		return nil
	}
}

// TODO does not stop on ctx cancel!!!!
func (ws *WS) Consume(ctx context.Context) (err error) {
	defer logrus.Debugln("ws.Start() ends")
	logrus.Debugln("ws.Start() call")

	for {
		_, message, err := ws.conn.ReadMessage()
		if err != nil {
			logrus.Errorln("Could not read from ws:", ws.url, err)
			return err
		}
		ws.reader <- message
	}
}

func (ws *WS) Write(msg []byte) error {
	err := ws.conn.WriteMessage(ws.msgType, msg)
	//logging.Logger.Errorln("Cannot write to ws", ws.url, err)
	return err
}

func (ws *WS) Close() {
	defer logrus.Debugln("ws.Close() ends")
	logrus.Debugln("ws.Close() call")
	//if ws.conn == nil {
	// if we closed session before even got the connection
	//return
	//}
	logrus.Warningln("Closing ws: ", ws.url)
	err := ws.conn.Close()
	if err != nil {
		logrus.Errorln("Could not close ws", ws.url, err)
	}
	close(ws.reader)
}
