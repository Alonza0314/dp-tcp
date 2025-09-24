package model

type ClientConfig struct {
	ClientIE ClientIE `yaml:"client" valid:"required"`
	LoggerIE LoggerIE `yaml:"logger" valid:"required"`
}

type ClientIE struct {
	TCP1DialAddr string `yaml:"tcp1_dial_addr" valid:"required"`
	TCP1DialPort int    `yaml:"tcp1_dial_port" valid:"required"`
	TCP1ConnAddr string `yaml:"tcp1_conn_addr" valid:"required"`
	TCP1ConnPort int    `yaml:"tcp1_conn_port" valid:"required"`
	TCP2DialAddr string `yaml:"tcp2_dial_addr" valid:"required"`
	TCP2DialPort int    `yaml:"tcp2_dial_port" valid:"required"`
	TCP2ConnAddr string `yaml:"tcp2_conn_addr" valid:"required"`
	TCP2ConnPort int    `yaml:"tcp2_conn_port" valid:"required"`

	TunnelDevice TunnelDevice `yaml:"tunnel_device" valid:"required"`
}
