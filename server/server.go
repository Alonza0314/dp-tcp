package server

import (
	"context"
	"errors"
	"io"
	"net"

	"github.com/Alonza0314/dp-tcp/constant"
	"github.com/Alonza0314/dp-tcp/logger"
	"github.com/Alonza0314/dp-tcp/model"
	"github.com/Alonza0314/dp-tcp/tun"
	"github.com/cespare/xxhash/v2"
	"github.com/cornelk/hashmap"
	"github.com/songgao/water"
)

type DpTcpServer struct {
	tcpServer1 *tcpServer
	tcpServer2 *tcpServer

	tunnelDeviceName string
	tunnelDeviceIP   string

	tunnelDevice *water.Interface

	readFromTun  chan []byte
	readFromTcp1 chan []byte
	readFromTcp2 chan []byte

	writeToTun chan []byte

	packetMap *hashmap.Map[uint64, struct{}]

	*logger.ServerLogger
}

func NewDpTcpServer(config *model.ServerConfig, serverLogger *logger.ServerLogger) *DpTcpServer {
	return &DpTcpServer{
		tcpServer1: newTcpServer(config.ServerIE.TCP1ListenAddr, config.ServerIE.TCP1ListenPort),
		tcpServer2: newTcpServer(config.ServerIE.TCP2ListenAddr, config.ServerIE.TCP2ListenPort),

		tunnelDeviceName: config.ServerIE.TunnelDeviceName,
		tunnelDeviceIP:   config.ServerIE.TunnelDeviceIP,

		readFromTun:  make(chan []byte),
		readFromTcp1: make(chan []byte),
		readFromTcp2: make(chan []byte),

		writeToTun: make(chan []byte),

		packetMap: hashmap.New[uint64, struct{}](),

		ServerLogger: serverLogger,
	}
}

func (s *DpTcpServer) Start(ctx context.Context) error {
	s.ServerLog.Infof("DpTcpServer starting...")

	if err := s.tcpServer1.listen(); err != nil {
		s.Tcp1Log.Errorf("TCP 1 server listen failed: %v", err)
		return err
	}
	s.Tcp1Log.Infof("TCP 1 server started at %s:%d", s.tcpServer1.listenAddr, s.tcpServer1.listenPort)

	if err := s.tcpServer2.listen(); err != nil {
		s.Tcp2Log.Errorf("TCP 2 server listen failed: %v", err)
		if err := s.tcpServer1.close(); err != nil {
			s.Tcp1Log.Errorf("TCP 1 server close failed: %v", err)
		}
		return err
	}
	s.Tcp2Log.Infof("TCP 2 server started at %s:%d", s.tcpServer2.listenAddr, s.tcpServer2.listenPort)

	go func() {
		if err := s.tcpServer1.accept(); err != nil {
			if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) {
				return
			}
			s.Tcp1Log.Errorf("TCP 1 server accept failed: %v", err)
			return
		}
		for {
			buffer := make([]byte, constant.BUFFER_SIZE)
			if n, err := s.tcpServer1.read(buffer); err != nil {
				s.Tcp1Log.Errorf("TCP 1 server read failed: %v", err)
				return
			} else {
				s.readFromTcp1 <- buffer[:n]
				s.Tcp1Log.Debugf("Received packet %d", xxhash.Sum64(buffer[:n]))
				s.Tcp1Log.Tracef("Received packet %x", buffer[:n])
			}
		}
	}()

	go func() {
		if err := s.tcpServer2.accept(); err != nil {
			if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) {
				return
			}
			s.Tcp2Log.Errorf("TCP 2 server accept failed: %v", err)
			return
		}
		for {
			buffer := make([]byte, constant.BUFFER_SIZE)
			if n, err := s.tcpServer2.read(buffer); err != nil {
				s.Tcp2Log.Errorf("TCP 2 server read failed: %v", err)
				return
			} else {
				s.readFromTcp2 <- buffer[:n]
				s.Tcp2Log.Debugf("Received packet %d", xxhash.Sum64(buffer[:n]))
				s.Tcp2Log.Tracef("Received packet %x", buffer[:n])
			}
		}
	}()

	if err := s.setupTunnelDevice(); err != nil {
		s.ServerLog.Errorf("Tunnel device setup failed: %v", err)
		if err := s.tcpServer1.close(); err != nil {
			s.Tcp1Log.Errorf("TCP 1 server close failed: %v", err)
		}
		if err := s.tcpServer2.close(); err != nil {
			s.Tcp2Log.Errorf("TCP 2 server close failed: %v", err)
		}
		return err
	}

	go s.packetDuplicate(ctx)
	go s.packetEliminate(ctx)

	s.ServerLog.Infof("DpTcpServer started")
	return nil
}

func (s *DpTcpServer) Stop() {
	s.ServerLog.Infof("DpTcpServer stopping...")

	close(s.readFromTun)
	close(s.readFromTcp1)
	close(s.readFromTcp2)

	if err := s.cleanUpTunnelDevice(); err != nil {
		s.TunLog.Errorf("Tunnel device cleanup failed: %v", err)
	}

	if err := s.tcpServer2.close(); err != nil {
		s.Tcp2Log.Errorf("TCP 2 server close failed: %v", err)
	}
	s.Tcp2Log.Infof("TCP 2 server stopped")

	if err := s.tcpServer1.close(); err != nil {
		s.Tcp1Log.Errorf("TCP 1 server close failed: %v", err)
	}
	s.Tcp1Log.Infof("TCP 1 server stopped")

	s.ServerLog.Infof("DpTcpServer stopped")
}

func (s *DpTcpServer) setupTunnelDevice() error {
	s.TunLog.Infof("Setting up tunnel device %s with IP %s", s.tunnelDeviceName, s.tunnelDeviceIP)

	tun, err := tun.BringUpUeTunnelDevice(s.tunnelDeviceName, s.tunnelDeviceIP)
	if err != nil {
		return err
	}
	s.tunnelDevice = tun

	// go routine to read from tunnel device
	go func() {
		for {
			buffer := make([]byte, constant.BUFFER_SIZE)
			n, err := s.tunnelDevice.Read(buffer)
			if err != nil {
				s.TunLog.Errorf("Error reading from tunnel device: %v", err)
				return
			}
			version := buffer[0] >> 4
			if version == 6 {
				continue
			}
			data := make([]byte, n)
			copy(data, buffer[:n])
			s.readFromTun <- data
		}
	}()

	// go routine to write to tunnel device
	go func() {
		for {
			data := <-s.writeToTun
			if _, err := s.tunnelDevice.Write(data); err != nil {
				s.TunLog.Errorf("Error writing to tunnel device: %v", err)
				return
			}
		}
	}()

	s.TunLog.Infof("Tunnel device %s with IP %s set up", s.tunnelDeviceName, s.tunnelDeviceIP)
	return nil
}

func (s *DpTcpServer) cleanUpTunnelDevice() error {
	s.TunLog.Infof("Cleaning up tunnel device %s", s.tunnelDeviceName)

	if err := tun.BringDownUeTunnelDevice(s.tunnelDeviceName); err != nil {
		return err
	}

	s.TunLog.Infof("Tunnel device %s cleaned up", s.tunnelDeviceName)
	return nil
}

func (s *DpTcpServer) packetDuplicate(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case data := <-s.readFromTun:
			data1 := make([]byte, len(data))
			copy(data1, data)
			data2 := make([]byte, len(data))
			copy(data2, data)

			go func() {
				if n, err := s.tcpServer1.write(data1); err != nil {
					s.TunLog.Errorf("Error writing to TCP 1 server: %v", err)
				} else {
					s.TunLog.Debugf("Wrote %d bytes to TCP 1 server", n)
				}

			}()
			go func() {
				if n, err := s.tcpServer2.write(data2); err != nil {
					s.TunLog.Errorf("Error writing to TCP 2 server: %v", err)
				} else {
					s.TunLog.Debugf("Wrote %d bytes to TCP 2 server", n)
				}
			}()
		}
	}
}

func (s *DpTcpServer) packetEliminate(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case packet := <-s.readFromTcp1:
			go s.packetEliminateMain(packet)
		case packet := <-s.readFromTcp2:
			go s.packetEliminateMain(packet)
		}
	}
}

func (s *DpTcpServer) packetEliminateMain(packet []byte) {
	h := xxhash.Sum64(packet)
	if _, ok := s.packetMap.Get(h); ok {
		s.writeToTun <- packet
		s.TunLog.Debugf("Eliminated packet %d", h)
		s.TunLog.Tracef("Eliminated packet %d, %x", h, packet)
		return
	}
	s.packetMap.Set(h, struct{}{})
	s.TunLog.Debugf("Packet %d stored", h)
	s.TunLog.Tracef("Packet %d stored, %x", h, packet)
}
