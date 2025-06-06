package config

type Config struct {
	epMasterConf *ServerEndpointConfig
	bridgePort   string
}

type ServerEndpointConfig struct {
	name string
	ip   string
	port string
	host string
}

var config *Config

func NewServerEndpointConf(name, ip, port string) *ServerEndpointConfig {
	return &ServerEndpointConfig{
		name: name,
		ip:   ip,
		port: port,
		host: ip + ":" + port,
	}
}

func NewConfig(epMasterConf *ServerEndpointConfig, bridgePort string) {
	config = &Config{
		epMasterConf: epMasterConf,
		bridgePort:   bridgePort,
	}
}

func GetMasterEndPoint() string {
	return config.epMasterConf.host
}

func GetBridgePort() string {
	return config.bridgePort
}
