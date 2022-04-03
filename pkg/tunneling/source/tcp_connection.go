package source

import (
	"context"
	"errors"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"time"
)

type TCPConnection struct {
	conn   net.Conn
	reader chan []byte
	logger *logrus.Logger
}

func NewTCPConnection(conn net.Conn, logger *logrus.Logger) *TCPConnection {
	return &TCPConnection{
		conn:   conn,
		reader: make(chan []byte),
		logger: logger,
	}
}

func (tcp *TCPConnection) GetReader() chan []byte {
	return tcp.reader
}

func (tcp *TCPConnection) Consume(ctx context.Context) (err error) {
	defer tcp.logger.Debugln("tcpConn.Consume() ends")
	tcp.logger.Debugln("tcpCOn.Consume()")
	for {
		select {
		case <-ctx.Done():
			tcp.Close()
			return
		default:
			err := tcp.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
			if err != nil {
				panic("cannot set deadline to tcpConn ")
			}
			buf := make([]byte, 1024)
			n, err := tcp.conn.Read(buf)
			if err != nil {
				if err == io.EOF {
					return nil
				}
				netErr, ok := err.(net.Error)
				isTimeout := ok && netErr.Timeout()
				if isTimeout {
					continue
				}
				return errors.New("Cannot read from tcp connection: " + err.Error())
			}
			message := buf[:n]
			tcp.reader <- message
		}
	}
}

func (tcp *TCPConnection) Close() {
	defer tcp.logger.Debugln("tcpConn.Close() ends")
	err := tcp.conn.Close()
	if err != nil {
		tcp.logger.Errorln("Could not close connection to tcpConn:", err)
	}
	tcp.conn = nil
	close(tcp.reader)
}

func (tcp *TCPConnection) Write(msg []byte) (err error) {
	_, err = tcp.conn.Write(msg)
	return err
}
