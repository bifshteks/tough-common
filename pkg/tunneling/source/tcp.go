package source

import (
	"context"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"net"
	"time"
)

type TCP struct {
	url    string
	conn   *net.TCPConn
	reader chan []byte
	logger *logrus.Logger
}

func NewTCP(url string, logger *logrus.Logger) *TCP {
	return &TCP{
		url:    url,
		conn:   nil,
		reader: make(chan []byte),
		logger: logger,
	}
}

func (tcp *TCP) GetUrl() string {
	return tcp.url
}

func (tcp *TCP) GetReader() chan []byte {
	return tcp.reader
}

func (tcp *TCP) Connect(ctx context.Context) error {
	tcp.logger.Debugf("tcp.Connect() on %s", tcp.url)
	tcp.logger.Infof("Connecting to tcp on %s", tcp.url)
	dialer := net.Dialer{Timeout: 5 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", tcp.url)
	if err != nil {
		errMsg := fmt.Sprintf("tcp dial failed: %s", err.Error())
		return NewFatalConnectError(errors.New(errMsg))
	}
	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		panic("cannot convert to tcpConn")
	}
	tcp.conn = tcpConn
	tcp.logger.Infof("Connected to tcp on %s", tcp.url)
	go func() {
		<-ctx.Done()
		tcp.Close()
	}()
	return nil
}

func (tcp *TCP) Consume(ctx context.Context) error {
	defer tcp.logger.Debugln("tcp.Consume() ends")
	tcp.logger.Debugln("tcp.Consume() call")

	// don't need to catch context done - we already created a goroutine in .Connect() method
	// waiting for that
	for {
		buffer := make([]byte, 1024)
		n, err := tcp.conn.Read(buffer)
		if err != nil {
			if IsClosedConnError(err) {
				return nil
			}
			errMsg := fmt.Sprintf(
				"Could not read from tcp on %s: %s", tcp.url, err,
			)
			return errors.New(errMsg)
		}
		tcp.reader <- buffer[:n]
	}
}

func (tcp *TCP) Write(msg []byte) error {
	_, err := tcp.conn.Write(msg)
	return err
}

func (tcp *TCP) Close() {
	defer tcp.logger.Debugln("tcp.Close() ends")
	tcp.logger.Debugln("tcp.Close() call")
	err := tcp.conn.Close()
	if err != nil {
		tcp.logger.Errorln("Could not close connection to tcp:", err)
	}
	tcp.conn = nil
	close(tcp.reader)
}
