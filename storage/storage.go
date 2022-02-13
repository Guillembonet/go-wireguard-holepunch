package storage

import (
	"fmt"
	"log"

	"github.com/guillembonet/go-wireguard-holepunch/communication/server"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type Storage struct {
	peers map[string]server.Peer
}

func NewStorage() *Storage {
	return &Storage{
		peers: make(map[string]server.Peer),
	}
}

func (s *Storage) UpsertPeer(publicKey, ip string) error {
	s.peers[publicKey] = server.Peer{
		Ip: ip,
	}
	log.Printf("added peer %s with ip: %s", publicKey, ip)
	return nil
}

func (s *Storage) GetPeer(publicKey string) (server.Peer, error) {
	peer, ok := s.peers[publicKey]
	if !ok {
		return server.Peer{}, fmt.Errorf("peer not found with public key: %s", publicKey)
	}
	log.Printf("retrieved peer %s with ip: %s", publicKey, peer.Ip)
	return peer, nil
}

func (s *Storage) ListAnnouncements() ([]server.Announcement, error) {
	res := []server.Announcement{}
	for peerKey := range s.peers {
		publicKey, err := wgtypes.ParseKey(peerKey)
		if err != nil {
			return nil, err
		}
		peer, ok := s.peers[peerKey]
		if !ok {
			return nil, fmt.Errorf("peer not found with public key: %s", publicKey)
		}
		res = append(res, server.Announcement{
			PublicKey: publicKey,
			Peer:      peer,
		})
	}
	return res, nil
}
