package storage

import (
	"fmt"
	"log"

	"github.com/guillembonet/go-wireguard-udpholepunch/communication/server"
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
