package endpoints

import (
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/guillembonet/go-wireguard-holepunch/communication/client"
)

// AnnounceEndpoint represents the ping endpoint
type AnnounceEndpoint struct {
	client     *client.Client
	serverIP   string
	serverPort int
}

// NewAnnounceEndpoint returns a new announce endpoint.
func NewAnnounceEndpoint(client *client.Client, serverIP string, serverPort int) *AnnounceEndpoint {
	return &AnnounceEndpoint{
		client:     client,
		serverIP:   serverIP,
		serverPort: serverPort,
	}
}

// RegisterRoutes registers the ping route
func (ae *AnnounceEndpoint) RegisterRoutes(r gin.IRoutes) {
	r.POST("/announce", ae.AnnounceHandler)
}

type AnnounceReq struct {
	ServerIp    string `json:"server_ip"`
	ServerPort  int    `json:"server_port"`
	InterfaceIP string `json:"interface_ip"`
}

func (ae *AnnounceEndpoint) AnnounceHandler(c *gin.Context) {
	var req AnnounceReq
	if err := c.BindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, NewErrorValidation("unparsable json", err))
		return
	}
	serverIp := net.ParseIP(req.ServerIp)
	if serverIp == nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, NewErrorValidation("invalid server ip", nil))
		return
	}
	interfaceIp := net.ParseIP(req.InterfaceIP)
	if serverIp == nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, NewErrorValidation("invalid interface ip", nil))
		return
	}
	err := ae.client.Announce(serverIp, req.ServerPort, interfaceIp)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, NewGenericError("announce error", err))
		return
	}
	c.Status(http.StatusCreated)
}
