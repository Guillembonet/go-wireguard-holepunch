package params

import "flag"

type Generic struct {
	Port *int

	TunnelSlash24IP *string
}

func (g *Generic) Init() {
	g.Port = flag.Int("port", 2001, "port used by the client for the wireguard interface")

	g.TunnelSlash24IP = flag.String("tunnelSlash24IP", "10.1.0.0", "cidr of the tunnel network (example: 10.0.1.0)")
}
