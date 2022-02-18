package communication

import (
	"fmt"
	"log"
	"net"
)

//CreateUDPSocket creates a udp socket
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

//SendUDPMessage sends a udp message to the given address using a given connection
func SendUDPMessage(conn *net.UDPConn, message string, address net.UDPAddr) error {
	_, err := conn.WriteTo([]byte(message), &address)
	if err != nil {
		return err
	}
	return nil
}

//ReadUDP reads from a udp connection and calls the provided handling function on a new
//goroutine with the message content and origin address
func ReadUDP(conn *net.UDPConn, handleMessageFunc func(string, *net.UDPAddr) error) error {
	buf := make([]byte, 1024)
	for {
		rlen, originAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			return err
		}
		go handleMessageFunc(string(buf[0:rlen]), originAddr)
	}
}

//ReadUDPDebug the same as ReadUDP but with error logging
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
