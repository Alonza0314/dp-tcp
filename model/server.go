package model

type ServerConfig struct {
	ServerIE ServerIE `yaml:"server" valid:"required"`
	LoggerIE LoggerIE `yaml:"logger" valid:"required"`
}

type ServerIE struct {
	TCP1ListenAddr string `yaml:"tcp1_listen_addr" valid:"required"`
	TCP1ListenPort int    `yaml:"tcp1_listen_port" valid:"required"`
	TCP2ListenAddr string `yaml:"tcp2_listen_addr" valid:"required"`
	TCP2ListenPort int    `yaml:"tcp2_listen_port" valid:"required"`

	TunnelDeviceName string `yaml:"tunnel_device_name" valid:"required"`
	TunnelDeviceIP   string `yaml:"tunnel_device_ip" valid:"required"`
}

