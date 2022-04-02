package source

import (
	"context"
	"errors"
	"github.com/sirupsen/logrus"
	"net"
	"time"
)

type TCPConnection struct {
	conn   net.Conn
	reader chan []byte
}

func NewTCPConnection(conn net.Conn) *TCPConnection {
	return &TCPConnection{
		conn:   conn,
		reader: make(chan []byte),
	}
}

func (tcp *TCPConnection) GetReader() chan []byte {
	return tcp.reader
}

func (tcp *TCPConnection) Consume(ctx context.Context) (err error) {
	defer logrus.Debugln("gorounting tcpConn.Start() ends")
	logrus.Debugln("start tcpCOn")
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
			if err == nil {
				message := buf[:n]
				tcp.reader <- message
				continue
			}
			netErr, ok := err.(net.Error)
			isTimeout := ok && netErr.Timeout()
			if isTimeout {
				continue
			}
			// todo check if it's really an error, not just end of connection
			return errors.New("Cannot read from tcp conn: " + err.Error())
		}
	}
}

func (tcp *TCPConnection) Close() {
	defer logrus.Debugln("tcpConn.Close() ends")
	if tcp.conn == nil {
		// if we closed session before even got the connection
		return
	}
	_ = tcp.conn.Close()
	close(tcp.reader)
}

func (tcp *TCPConnection) Write(msg []byte) (err error) {
	_, err = tcp.conn.Write(msg)
	return err
}
