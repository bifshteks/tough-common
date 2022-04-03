package source

import (
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type WSDialer interface {
	DialContext(ctx context.Context, url string, requestHeader http.Header)
}
type WS struct {
	url           string
	conn          *websocket.Conn
	reader        chan []byte
	msgType       int
	requestHeader http.Header
	logger        *logrus.Logger
}

func NewWS(url string, msgType int, requestHeader http.Header, logger *logrus.Logger) *WS {
	return &WS{
		url:           url,
		conn:          nil,
		reader:        make(chan []byte),
		msgType:       msgType,
		requestHeader: requestHeader,
		logger:        logger,
	}
}

func NewAuthorizedWS(url string, msgType int, headerName string, headerValue string, logger *logrus.Logger) *WS {
	header := http.Header{
		headerName: []string{headerValue},
	}
	return NewWS(url, msgType, header, logger)
}

func (ws *WS) GetUrl() string {
	return ws.url
}

func (ws *WS) GetReader() chan []byte {
	return ws.reader
}

func (ws *WS) Connect(ctx context.Context) (err error) {
	defer ws.logger.Infof("ws.Connect() on %s end", ws.url)
	ws.logger.Infof("ws.Connect() on %s", ws.url)
	// todo does dialContext close connection on ctx expiration as well ?
	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 10 * time.Second
	conn, resp, err := dialer.DialContext(ctx, ws.url, ws.requestHeader)
	if err != nil {
		fatalResponseError := resp != nil && (resp.StatusCode == http.StatusBadRequest ||
			resp.StatusCode == http.StatusUnauthorized)
		if fatalResponseError {
			errMsg := fmt.Sprintf(
				"ws connection cannot connect to %s, response status is %d",
				ws.url, resp.StatusCode,
			)
			return NewFatalConnectError(errors.New(errMsg))
		}
		errMsg := fmt.Sprintf("Cannot connect to ws %s: %s", ws.url, err)
		return errors.New(errMsg)
	}
	ws.conn = conn
	ws.logger.Infof("ws connected to %s", ws.url)
	// cannot set readDeadLine - https://github.com/gorilla/websocket/issues/474,
	// so use this goroutine
	go func() {
		<-ctx.Done()
		ws.Close()
	}()
	return nil
}

func (ws *WS) Consume(ctx context.Context) (err error) {
	defer ws.logger.Debugln("ws.Consume() ends")
	ws.logger.Debugf("ws.Consume() call %s", ws.url)

	// don't need to catch context done - we already created a goroutine in .Connect() method
	// waiting for that
	for {
		_, message, err := ws.conn.ReadMessage()
		if err != nil {
			if IsClosedConnError(err) {
				return nil
			}
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
	defer ws.logger.Debugln("ws.Close() ends")
	ws.logger.Debugf("ws.Close() call for ws on %s", ws.url)
	err := ws.conn.Close()
	if err != nil {
		ws.logger.Errorf("Could not close ws on %s: %s", ws.url, err)
	}
	ws.conn = nil
	close(ws.reader)
}
