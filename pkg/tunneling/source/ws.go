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
	logrus.Debugf("ws.Connect() on %s", ws.url)
	select {
	case <-ctx.Done():
		return nil
	default:
		conn, resp, err := websocket.DefaultDialer.DialContext(ctx, ws.url, ws.requestHeader)
		if resp != nil {
			responseWithKnownBadStatus := resp.StatusCode == http.StatusBadRequest ||
				resp.StatusCode == http.StatusUnauthorized
			if responseWithKnownBadStatus {
				return errors.New(
					fmt.Sprintf(
						"ws connection cannot connect to %s, response status is %d",
						ws.url, resp.StatusCode,
					),
				)
			}
		}
		if err != nil {
			return errors.New(fmt.Sprintf("Cannot connect to ws %s: %s", ws.url, err))
		}
		ws.conn = conn
		logrus.Infof("ws connected to %s", ws.url)
		// cannot set readDeadLine - https://github.com/gorilla/websocket/issues/474,
		// so use this goroutine
		go func() {
			<-ctx.Done()
			ws.Close()
		}()
		return nil
	}
}

func (ws *WS) Consume(ctx context.Context) (err error) {
	defer logrus.Debugln("ws.Start() ends")
	logrus.Debugf("ws.Start() call %s", ws.url)

	// don't need to catch context done - we already created a goroutine in .Connect() method
	// waiting for that
	for {
		_, message, err := ws.conn.ReadMessage()
		if err != nil {
			normalClosure := websocket.IsCloseError(err, websocket.CloseNormalClosure)
			if normalClosure {
				return nil
			}
			return errors.New(fmt.Sprintf("Cound not read from ws on %s: %s", ws.url, err))
		}
		ws.reader <- message
	}
}

func (ws *WS) Write(msg []byte) error {
	return ws.conn.WriteMessage(ws.msgType, msg)
}

func (ws *WS) Close() {
	defer logrus.Debugln("ws.Close() ends")
	logrus.Debugf("ws.Close() call for ws on %s", ws.url)
	err := ws.conn.Close()
	if err != nil {
		logrus.Errorf("Could not close ws on %s: %s", ws.url, err)
	}
	ws.conn = nil
	close(ws.reader)
}
