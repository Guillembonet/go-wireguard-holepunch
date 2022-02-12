package params

import "flag"

type Client struct {
	ServerIP               *string
	ServerPort             *int
	WireguardInterfaceName *string
}

func (c *Client) Init() {
	c.ServerIP = flag.String("serverIp", "empty", "IP address of the server")
	c.ServerPort = flag.Int("serverPort", 2001, "port used by the server for udp communication")
	c.WireguardInterfaceName = flag.String("wireguardInterfaceName", "wg0", "name of the wireguard interface")
}
