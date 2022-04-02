package source

import (
	"context"
	"gitlab.com/hi-it/tough-connect/tough_client/internal/logging"
	"net"
	"time"
)

type TCP struct {
	url    string
	conn   *net.TCPConn
	reader chan []byte
}

func NewTCP(url string) *TCP {
	return &TCP{
		url:    url,
		conn:   nil,
		reader: make(chan []byte),
	}
}

func (tcp *TCP) GetUrl() string {
	return tcp.url
}

func (tcp *TCP) GetReader() chan []byte {
	return tcp.reader
}

func (tcp *TCP) Connect(ctx context.Context) error {
	logging.Logger.Debugln("tcp.Connect()")
	select {
	case <-ctx.Done():
		return nil
	default:
		logging.Logger.Infoln("Connecting to VNC on", tcp.url)
		conn, err := net.DialTimeout("tcp", tcp.url, 3*time.Second)
		if err != nil {
			logging.Logger.Errorln("VNC dial failed:", err)
			return err
		}
		tcpConn, ok := conn.(*net.TCPConn)
		if !ok {
			panic("cannot convert to tcpConn")
		}
		tcp.conn = tcpConn
		logging.Logger.Infoln("Connected to vnc")
		go func() {
			<-ctx.Done()
			tcp.Close()
		}()
		return nil
	}
}

func (tcp *TCP) Start(ctx context.Context) error {
	defer logging.Logger.Debugln("tcp.Start() ends")
	if tcp.conn == nil {
		panic("try to start tcp source before connection is created")
	}
	logging.Logger.Debugln("tcp.Start() call")
	for {
		select {
		case <-ctx.Done():
			logging.Logger.Debugln("tcp.Start() ctx case")
			return nil
		default:
			err := tcp.conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			if err != nil {
				panic("cannot set deadline to tcpConn ")
			}
			buffer := make([]byte, 1024)
			n, err := tcp.conn.Read(buffer)
			if err == nil { // when connection is closed vnc sends 1 byte, dk why
				tcp.reader <- buffer[:n]
				continue
			}
			netErr, ok := err.(net.Error)
			isTimeout := ok && netErr.Timeout()
			if isTimeout {
				continue
			}
			logging.Logger.Errorln("Could not read from tcp", err, n, string(buffer), ".", buffer[:10])
			return err
		}
	}
}

func (tcp *TCP) Write(msg []byte) error {
	_, err := tcp.conn.Write(msg)
	if err != nil {
		logging.Logger.Errorln("Cannot write to tcp", err)
		return err
	}
	return nil
}

func (tcp *TCP) Close() {
	defer logging.Logger.Debugln("tcp.Close() ends")
	logging.Logger.Debugln("tcp.Close() call")
	err := tcp.conn.Close()
	if err != nil {
		logging.Logger.Errorln("Could not close connection to tcp:", err)
		return
	}
	// waiting (with timeout) for the server to close the connection.
	<-time.After(time.Second)
	close(tcp.reader)
}
