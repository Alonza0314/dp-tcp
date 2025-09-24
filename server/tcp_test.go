package server

import (
	"net"
	"testing"
	"time"
)

func TestTcpServer(t *testing.T) {
	passFlag := true

	server := newTcpServer("127.0.0.1", 31413)
	defer server.close()

	if err := server.listen(); err != nil {
		t.Fatalf("listen failed: %v", err)
	}

	go func(t *testing.T) {
		if err := server.accept(); err != nil {
			t.Errorf("accept failed: %v", err)
			passFlag = false
		}

		buf := make([]byte, 1024)
		if n, err := server.read(buf); err != nil {
			t.Errorf("read failed: %v", err)
		} else {
			if string(buf[:n]) != "test TCP" {
				t.Errorf("expected 'test TCP', got '%s'", string(buf[:n]))
				passFlag = false
			}
		}

		if n, err := server.write([]byte("test TCP")); err != nil {
			t.Errorf("write failed: %v", err)
		} else {
			if n != len([]byte("test TCP")) {
				t.Errorf("expected 'test TCP', got '%s'", string(buf[:n]))
				passFlag = false
			}
		}
	}(t)

	time.Sleep(1 * time.Second)

	conn, err := net.Dial("tcp", "127.0.0.1:31413")
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

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

	if !passFlag {
		t.Fatalf("test failed")
	}
}
