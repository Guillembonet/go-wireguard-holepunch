package endpoints

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/guillembonet/go-wireguard-holepunch/communication/server"
	"github.com/guillembonet/go-wireguard-holepunch/storage"
)

// AnnounceEndpoint represents the ping endpoint
type ListEndpoint struct {
	storage *storage.Storage
}

// NewListEndpoint returns a new announce endpoint.
func NewListEndpoint(storage *storage.Storage) *ListEndpoint {
	return &ListEndpoint{
		storage: storage,
	}
}

// RegisterRoutes registers the ping route
func (le *ListEndpoint) RegisterRoutes(r gin.IRoutes) {
	r.GET("/list", le.ListHandler)
}

type AnnouncementResp struct {
	PublicKey string `json:"public_key"`
	CIDR      string `json:"cidr"`
}

type ListResp []AnnouncementResp

func (le *ListEndpoint) ListHandler(c *gin.Context) {
	peers, err := le.storage.ListAnnouncements()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, NewGenericError("getting list failed", err))
		return
	}
	c.JSON(http.StatusOK, toListReponse(peers))
}

func toListReponse(peers []server.Announcement) []AnnouncementResp {
	res := []AnnouncementResp{}
	for _, peer := range peers {
		res = append(res, AnnouncementResp{
			PublicKey: peer.PublicKey.String(),
			CIDR:      peer.Peer.Ip,
		})
	}
	return res
}
