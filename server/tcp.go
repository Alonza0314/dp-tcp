package server

import (
	"fmt"
	"net"
)

type tcpServer struct {
	listenAddr string
	listenPort int

	listener net.Listener

	acceptConn net.Conn
}

func newTcpServer(listenAddr string, listenPort int) *tcpServer {
	return &tcpServer{
		listenAddr: listenAddr,
		listenPort: listenPort,

		listener: nil,

		acceptConn: nil,
	}
}

func (s *tcpServer) listen() error {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.listenAddr, s.listenPort))
	if err != nil {
		return err
	}
	s.listener = listener
	return nil
}

func (s *tcpServer) accept() error {
	conn, err := s.listener.Accept()
	if err != nil {
		return err
	}
	s.acceptConn = conn
	return nil
}

func (s *tcpServer) close() error {
	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			return err
		}
	}
	if s.acceptConn != nil {
		if err := s.acceptConn.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (s *tcpServer) read(buf []byte) (n int, err error) {
	return s.acceptConn.Read(buf)
}

func (s *tcpServer) write(data []byte) (n int, err error) {
	return s.acceptConn.Write(data)
}
