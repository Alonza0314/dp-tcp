package client

import (
	"net"
	"testing"
	"time"
)

func TestTcpClient(t *testing.T) {
	server, err := net.Listen("tcp", "127.0.0.1:31414")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	defer func() {
		if err := server.Close(); err != nil {
			t.Errorf("close failed: %v", err)
		}
	}()

	go func() {
		conn, err := server.Accept()
		if err != nil {
			t.Errorf("accept failed: %v", err)
		}
		defer func() {
			if err := conn.Close(); err != nil {
				t.Errorf("close failed: %v", err)
			}
		}()

		buf := make([]byte, 1024)
		if n, err := conn.Read(buf); err != nil {
			t.Errorf("read failed: %v", err)
		} else {
			if string(buf[:n]) != "test TCP" {
				t.Errorf("expected 'test TCP', got '%s'", string(buf[:n]))
			}
		}

		if n, err := conn.Write([]byte("test TCP")); err != nil {
			t.Errorf("write failed: %v", err)
		} else {
			if n != len([]byte("test TCP")) {
				t.Errorf("expected 'test TCP', got '%s'", string(buf[:n]))
			}
		}
	}()

	time.Sleep(1 * time.Second)

	client := newTcpClient("127.0.0.1", 31414, "127.0.0.1", 31415)
	defer func() {
		if err := client.close(); err != nil {
			t.Errorf("close failed: %v", err)
		}
	}()

	if err := client.dial(); err != nil {
		t.Fatalf("connect failed: %v", err)
	}

	if _, err := client.write([]byte("test TCP")); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	time.Sleep(1 * time.Second)

	buf := make([]byte, 1024)
	if n, err := client.read(buf); err != nil {
		t.Fatalf("read failed: %v", err)
	} else {
		if string(buf[:n]) != "test TCP" {
			t.Fatalf("expected 'test TCP', got '%s'", string(buf[:n]))
		}
	}
}
