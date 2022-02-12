package communication

import (
	"fmt"
	"log"
	"net"
)

func CreateUDPSocket(port string) (*net.UDPConn, error) {
	addr, err := net.ResolveUDPAddr("udp", port)
	if err != nil {
		return nil, err
	}
	sock, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}
	return sock, nil
}

func SendUDPMessage(conn *net.UDPConn, message string, address net.UDPAddr) error {
	_, err := conn.WriteTo([]byte(message), &address)
	if err != nil {
		return err
	}
	return nil
}

func ReadUDP(conn *net.UDPConn, handleMessage func(string, *net.UDPAddr) error) error {
	buf := make([]byte, 1024)
	for {
		rlen, originAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			return err
		}
		go handleMessage(string(buf[0:rlen]), originAddr)
	}
}

func ReadUDPDebug(conn *net.UDPConn, handleMessage func(string, *net.UDPAddr) error) error {
	buf := make([]byte, 1024)
	for {
		rlen, originAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			return err
		}
		go func() {
			err := handleMessage(string(buf[0:rlen]), originAddr)
			if err != nil {
				log.Println(fmt.Errorf("handle message failed: %w", err))
			}
		}()
	}
}
