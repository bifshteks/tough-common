package source

import (
	"context"
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

func (tcp *TCPConnection) GetName() string {
	return "TCPConnection"
}

func (tcp *TCPConnection) GetReader() chan []byte {
	return tcp.reader
}
func (tcp *TCPConnection) Connect(ctx context.Context, cancelFunc context.CancelFunc) {}

func (tcp *TCPConnection) Start(ctx context.Context, cancel context.CancelFunc) {
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
			logrus.Errorln("Cannot read from tcp conn", err)
			cancel()
			return
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

func (tcp *TCPConnection) Write(msg []byte) {
	_, err := tcp.conn.Write(msg)
	if err != nil {
		logrus.Errorln("Could not write to tcp connection", err, msg)
	}
}
