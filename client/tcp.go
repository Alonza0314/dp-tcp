package client

import (
	"fmt"
	"net"
)

type tcpClient struct {
	dialAddr string
	dialPort int

	connAddr string
	connPort int

	conn net.Conn
}

func newTcpClient(dialAddr string, dialPort int, connAddr string, connPort int) *tcpClient {
	return &tcpClient{
		dialAddr: dialAddr,
		dialPort: dialPort,

		connAddr: connAddr,
		connPort: connPort,

		conn: nil,
	}
}

func (c *tcpClient) dial() error {
	remoteAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", c.dialAddr, c.dialPort))
	if err != nil {
		return err
	}

	localAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", c.connAddr, c.connPort))
	if err != nil {
		return err
	}

	dialer := &net.Dialer{
		LocalAddr: localAddr,
	}

	conn, err := dialer.Dial("tcp", remoteAddr.String())
	if err != nil {
		return err
	}

	c.conn = conn
	return nil
}

func (c *tcpClient) close() error {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (c *tcpClient) read(buf []byte) (n int, err error) {
	return c.conn.Read(buf)
}

func (c *tcpClient) write(data []byte) (n int, err error) {
	return c.conn.Write(data)
}
