package endpoints

import (
	"fmt"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/guillembonet/go-wireguard-udpholepunch/communication/client"
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

// Ping represents the ping response
// swagger:response Ping
type Message struct {
	Message string `json:"message"`
}

func (ae *AnnounceEndpoint) AnnounceHandler(context *gin.Context) {
	err := ae.client.Announce(net.ParseIP(ae.serverIP), ae.serverPort, "15.13.10.1")
	if err != nil {
		fmt.Println(err)
		context.JSON(http.StatusInternalServerError, err)
		return
	}
	context.Status(http.StatusOK)
}
