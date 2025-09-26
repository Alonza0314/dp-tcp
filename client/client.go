package client

import (
	"context"
	"errors"
	"io"
	"net"
	"sync"

	"github.com/Alonza0314/dp-tcp/constant"
	"github.com/Alonza0314/dp-tcp/logger"
	"github.com/Alonza0314/dp-tcp/model"
	"github.com/Alonza0314/dp-tcp/tun"
	"github.com/cespare/xxhash/v2"
	"github.com/cornelk/hashmap"
	"github.com/songgao/water"
)

type DpTcpClient struct {
	tcpClient1 *tcpClient
	tcpClient2 *tcpClient

	tunnelDeviceName string
	tunnelDeviceIP   string

	tunnelDevice *water.Interface

	readFromTun  chan []byte
	readFromTcp1 chan []byte
	readFromTcp2 chan []byte

	writeToTun chan []byte

	packetMap *hashmap.Map[uint64, struct{}]

	*logger.ClientLogger
}

func NewDpTcpClient(config *model.ClientConfig, clientLogger *logger.ClientLogger) *DpTcpClient {
	return &DpTcpClient{
		tcpClient1: newTcpClient(config.ClientIE.TCP1DialAddr, config.ClientIE.TCP1DialPort, config.ClientIE.TCP1ConnAddr, config.ClientIE.TCP1ConnPort),
		tcpClient2: newTcpClient(config.ClientIE.TCP2DialAddr, config.ClientIE.TCP2DialPort, config.ClientIE.TCP2ConnAddr, config.ClientIE.TCP2ConnPort),

		tunnelDeviceName: config.ClientIE.TunnelDevice.Name,
		tunnelDeviceIP:   config.ClientIE.TunnelDevice.IP,

		readFromTun:  make(chan []byte),
		readFromTcp1: make(chan []byte),
		readFromTcp2: make(chan []byte),

		writeToTun: make(chan []byte),

		packetMap: hashmap.New[uint64, struct{}](),

		ClientLogger: clientLogger,
	}
}

func (c *DpTcpClient) Start(ctx context.Context) error {
	c.ClientLog.Infof("DpTcpClient starting...")

	wg, dialSuccess := sync.WaitGroup{}, true
	wg.Add(2)

	go func() {
		defer wg.Done()
		if err := c.tcpClient1.dial(); err != nil {
			c.ClientLog.Errorf("TCP 1 client dial failed: %v", err)
			dialSuccess = false
			return
		}
		c.ClientLog.Infof("TCP 1 client dialed to %s:%d", c.tcpClient1.dialAddr, c.tcpClient1.dialPort)
	}()

	go func() {
		defer wg.Done()
		if err := c.tcpClient2.dial(); err != nil {
			c.ClientLog.Errorf("TCP 2 client dial failed: %v", err)
			dialSuccess = false
			return
		}
		c.ClientLog.Infof("TCP 2 client dialed to %s:%d", c.tcpClient2.dialAddr, c.tcpClient2.dialPort)
	}()

	wg.Wait()
	if !dialSuccess {
		return errors.New("dial failed")
	}

	go func() {
		for {
			buffer := make([]byte, constant.BUFFER_SIZE)
			if n, err := c.tcpClient1.read(buffer); err != nil {
				if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) {
					return
				}
				c.ClientLog.Errorf("TCP 1 client read failed: %v", err)
				return
			} else {
				c.readFromTcp1 <- buffer[:n]
				c.ClientLog.Debugf("Received packet %d", xxhash.Sum64(buffer[:n]))
				c.ClientLog.Tracef("Received packet %x", buffer[:n])
			}
		}
	}()

	go func() {
		for {
			buffer := make([]byte, constant.BUFFER_SIZE)
			if n, err := c.tcpClient2.read(buffer); err != nil {
				if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) {
					return
				}
				c.ClientLog.Errorf("TCP 2 client read failed: %v", err)
				return
			} else {
				c.readFromTcp2 <- buffer[:n]
				c.ClientLog.Debugf("Received packet %d", xxhash.Sum64(buffer[:n]))
				c.ClientLog.Tracef("Received packet %x", buffer[:n])
			}
		}
	}()

	if err := c.setupTunnelDevice(); err != nil {
		c.ClientLog.Errorf("Tunnel device setup failed: %v", err)
		if err := c.tcpClient1.close(); err != nil {
			c.ClientLog.Errorf("TCP 1 client close failed: %v", err)
		}
		if err := c.tcpClient2.close(); err != nil {
			c.ClientLog.Errorf("TCP 2 client close failed: %v", err)
		}
		return err
	}
	c.ClientLog.Infof("Tunnel device set up")

	go c.packetDuplicate(ctx)
	go c.packetEliminate(ctx)

	c.ClientLog.Infof("DpTcpClient started")
	return nil
}

func (c *DpTcpClient) Stop() {
	c.ClientLog.Infof("DpTcpClient stopping...")

	close(c.readFromTun)
	close(c.readFromTcp1)
	close(c.readFromTcp2)

	if err := c.cleanUpTunnelDevice(); err != nil {
		c.TunLog.Errorf("Tunnel device cleanup failed: %v", err)
	}

	if err := c.tcpClient2.close(); err != nil {
		c.Tcp2Log.Errorf("TCP 2 client close failed: %v", err)
	}
	c.Tcp2Log.Infof("TCP 2 client stopped")

	if err := c.tcpClient1.close(); err != nil {
		c.Tcp1Log.Errorf("TCP 1 client close failed: %v", err)
	}
	c.Tcp1Log.Infof("TCP 1 client stopped")

	c.ClientLog.Infof("DpTcpClient stopped")
}

func (c *DpTcpClient) setupTunnelDevice() error {
	c.TunLog.Infof("Setting up tunnel device %s with IP %s", c.tunnelDeviceName, c.tunnelDeviceIP)

	tun, err := tun.BringUpUeTunnelDevice(c.tunnelDeviceName, c.tunnelDeviceIP)
	if err != nil {
		return err
	}
	c.tunnelDevice = tun

	// go routine to read from tunnel device
	go func() {
		for {
			buffer := make([]byte, constant.BUFFER_SIZE)
			n, err := c.tunnelDevice.Read(buffer)
			if err != nil {
				c.TunLog.Errorf("Error reading from tunnel device: %v", err)
				return
			}
			version := buffer[0] >> 4
			if version == 6 {
				continue
			}
			data := make([]byte, n)
			copy(data, buffer[:n])
			c.readFromTun <- data
		}
	}()

	// go routine to write to tunnel device
	go func() {
		for {
			data := <-c.writeToTun
			if _, err := c.tunnelDevice.Write(data); err != nil {
				c.TunLog.Errorf("Error writing to tunnel device: %v", err)
				return
			}
		}
	}()

	c.TunLog.Infof("Tunnel device %s with IP %s set up", c.tunnelDeviceName, c.tunnelDeviceIP)
	return nil
}

func (c *DpTcpClient) cleanUpTunnelDevice() error {
	c.TunLog.Infof("Cleaning up tunnel device %s", c.tunnelDeviceName)

	if err := tun.BringDownUeTunnelDevice(c.tunnelDeviceName); err != nil {
		return err
	}

	c.TunLog.Infof("Tunnel device %s cleaned up", c.tunnelDeviceName)
	return nil
}

func (c *DpTcpClient) packetDuplicate(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case data := <-c.readFromTun:
			data1 := make([]byte, len(data))
			copy(data1, data)
			data2 := make([]byte, len(data))
			copy(data2, data)

			go func() {
				if n, err := c.tcpClient1.write(data1); err != nil {
					c.TunLog.Errorf("Error writing to TCP 1 server: %v", err)
				} else {
					c.TunLog.Debugf("Wrote %d bytes to TCP 1 server", n)
				}

			}()
			go func() {
				if n, err := c.tcpClient2.write(data2); err != nil {
					c.TunLog.Errorf("Error writing to TCP 2 server: %v", err)
				} else {
					c.TunLog.Debugf("Wrote %d bytes to TCP 2 server", n)
				}
			}()
		}
	}
}

func (c *DpTcpClient) packetEliminate(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case packet := <-c.readFromTcp1:
			go c.packetEliminateMain(packet)
		case packet := <-c.readFromTcp2:
			go c.packetEliminateMain(packet)
		}
	}
}

func (c *DpTcpClient) packetEliminateMain(packet []byte) {
	h := xxhash.Sum64(packet)
	if _, ok := c.packetMap.Get(h); ok {
		c.packetMap.Del(h)
		c.TunLog.Debugf("Eliminated packet %d", h)
		c.TunLog.Tracef("Eliminated packet %d, %x", h, packet)
		return
	}
	c.writeToTun <- packet
	c.packetMap.Set(h, struct{}{})
	c.TunLog.Debugf("Packet %d stored", h)
	c.TunLog.Tracef("Packet %d stored, %x", h, packet)
}
