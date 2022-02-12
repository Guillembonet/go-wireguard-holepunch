package constants

import "time"

const DEFAULT_BASE_IP = "10.1.0."
const DEFAULT_FIREWALL_MARK = 2349
const DEFAULT_KEEPALIVE = time.Second * 5

const (
	AnnounceQuery string = "a"
	GetQuery      string = "g"
	PeerQuery     string = "peer"
	HelpQuery     string = "help"
)
