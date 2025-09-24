package server

import (
	"net"
	"testing"
	"time"
)

func TestTcpServer(t *testing.T) {
	server := newTcpServer("127.0.0.1", 31413)
	defer func() {
		if err := server.close(); err != nil {
			t.Errorf("close failed: %v", err)
		}
	}()

	if err := server.listen(); err != nil {
		t.Fatalf("listen failed: %v", err)
	}

	go func(t *testing.T) {
		if err := server.accept(); err != nil {
			t.Errorf("accept failed: %v", err)
		}

		buf := make([]byte, 1024)
		if n, err := server.read(buf); err != nil {
			t.Errorf("read failed: %v", err)
		} else {
			if string(buf[:n]) != "test TCP" {
				t.Errorf("expected 'test TCP', got '%s'", string(buf[:n]))
			}
		}

		if n, err := server.write([]byte("test TCP")); err != nil {
			t.Errorf("write failed: %v", err)
		} else {
			if n != len([]byte("test TCP")) {
				t.Errorf("expected 'test TCP', got '%s'", string(buf[:n]))
			}
		}
	}(t)

	time.Sleep(1 * time.Second)

	conn, err := net.Dial("tcp", "127.0.0.1:31413")
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			t.Errorf("close failed: %v", err)
		}
	}()

	if _, err := conn.Write([]byte("test TCP")); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	time.Sleep(1 * time.Second)

	buf := make([]byte, 1024)
	if n, err := conn.Read(buf); err != nil {
		t.Fatalf("read failed: %v", err)
	} else {
		if string(buf[:n]) != "test TCP" {
			t.Fatalf("expected 'test TCP', got '%s'", string(buf[:n]))
		}
	}
}
