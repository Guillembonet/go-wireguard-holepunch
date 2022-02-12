package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/guillembonet/go-wireguard-holepunch/communication"
	"github.com/guillembonet/go-wireguard-holepunch/constants"
	"github.com/guillembonet/go-wireguard-holepunch/utils"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type Server struct {
	conn    *net.UDPConn
	lock    sync.Mutex
	storage Storage
}

type Storage interface {
	UpsertPeer(publicKey, ip string) error
	GetPeer(publicKey string) (Peer, error)
}

type Peer struct {
	Ip string
}

func NewServer(conn *net.UDPConn, storage Storage) *Server {
	return &Server{
		conn:    conn,
		storage: storage,
	}
}

func (s *Server) Start() error {
	return communication.ReadUDPDebug(s.conn, s.handlePacket)
}

func (s *Server) handlePacket(message string, originAddr *net.UDPAddr) error {
	args := utils.GetQuery(message)

	log.Printf((fmt.Sprintf("received message from %s with content: %s", originAddr.String(), message)))

	switch args[0] {
	// a <public_key> <ip>
	case constants.AnnounceQuery:
		publicKey, err := wgtypes.ParseKey(args[1])
		if err != nil {
			return fmt.Errorf("public key parsing failed: %w", err)
		}
		ip := args[2]
		err = s.storage.UpsertPeer(publicKey.String(), ip)
		if err != nil {
			return fmt.Errorf("saving peer failed: %w", err)
		}
		//Reply with add
		s.lock.Lock()
		defer s.lock.Unlock()
		err = communication.SendUDPMessage(s.conn, fmt.Sprintf("%s %s %s", constants.AnnounceQuery, publicKey.String(), ip), *originAddr)
		if err != nil {
			return fmt.Errorf("sending reply message failed: %w", err)
		}
		log.Printf("replied with: a %s %s", publicKey.String(), ip)
	// g <public_key>
	case constants.GetQuery:
		publicKey, err := wgtypes.ParseKey(args[1])
		if err != nil {
			return fmt.Errorf("public key parsing failed: %w", err)
		}
		peer, err := s.storage.GetPeer(publicKey.String())
		if err != nil {
			return fmt.Errorf("getting peer failed: %w", err)
		}
		jsonData, err := json.Marshal(peer)
		if err != nil {
			return fmt.Errorf("encoding peer failed: %w", err)
		}
		s.lock.Lock()
		defer s.lock.Unlock()
		err = communication.SendUDPMessage(s.conn, fmt.Sprintf("p %s", string(jsonData)), *originAddr)
		if err != nil {
			return fmt.Errorf("returning peer data failed: %w", err)
		}
		log.Printf("replied to %s with peer data: %s", originAddr.String(), string(jsonData))
	}
	return nil
}
