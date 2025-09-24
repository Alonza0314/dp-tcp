package server

import (
	"os"
	"testing"
)

var testUeTunnelDeviceName = []struct {
	name             string
	tunnelDeviceName string
	ip               string
}{
	{
		name:             "test1",
		tunnelDeviceName: "dptcptun",
		ip:               "10.0.0.1",
	},
}

func TestUeTunnelDeviceName(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("Skipping test because it requires root privileges")
	}
	for _, test := range testUeTunnelDeviceName {
		t.Run(test.name, func(t *testing.T) {
			_, err := bringUpUeTunnelDevice(test.tunnelDeviceName, test.ip)
			if err != nil {
				t.Fatalf("Error bringing up tunnel device: %v", err)
			}
			defer func() {
				if err := bringDownUeTunnelDevice(test.tunnelDeviceName); err != nil {
					t.Fatalf("Error bringing down tunnel device: %v", err)
				}
			}()

			t.Logf("Tunnel device %s brought up", test.tunnelDeviceName)
			t.Logf("Tunnel device %s IP: %s", test.tunnelDeviceName, test.ip)
		})
	}
}
