package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/routing"
	"github.com/guillembonet/go-wireguard-udpholepunch/connection"
	"github.com/guillembonet/go-wireguard-udpholepunch/constants"
	"github.com/guillembonet/go-wireguard-udpholepunch/utils"
	"github.com/mdlayher/arp"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type Client struct {
	port    int
	manager ConnectionManager
}

type ConnectionManager interface {
	SetPeer(publicKey wgtypes.Key, cidr string, endpoint *net.UDPAddr, keepalive *time.Duration) error
	GetPublicKey() (*wgtypes.Key, error)
	GetInterfaceIP() (net.Addr, error)
	SetInterfaceIP(ip string) error
}

func NewClient(port int, manager ConnectionManager) *Client {
	return &Client{
		port:    port,
		manager: manager,
	}
}

func (c *Client) Announce(serverIP net.IP, serverPort int, ip string) error {
	err := c.manager.SetInterfaceIP(ip)
	if err != nil {
		return fmt.Errorf("error setting interface ip: %w", err)
	}
	publicKey, err := c.manager.GetPublicKey()
	if err != nil {
		return err
	}
	payload := fmt.Sprintf("%s %s %s", constants.AnnounceQuery, publicKey.String(), ip)
	return c.sendMessage(serverIP, serverPort, payload, true)
}

func (c *Client) GetPeer(serverIP net.IP, serverPort int, peerPublicKey wgtypes.Key) error {
	payload := fmt.Sprintf("%s %s", constants.GetQuery, peerPublicKey.String())
	return c.sendMessage(serverIP, serverPort, payload, true)
}

func (c *Client) sendMessage(serverIP net.IP, serverPort int, message string, expectReply bool) error {
	router, err := routing.New()
	if err != nil {
		return fmt.Errorf("error while creating routing object: %w", err)
	}
	iface, gatewayIP, sourceIP, err := router.Route(serverIP)
	if err != nil {
		return fmt.Errorf("error routing to ip %s: %w", serverIP, err)
	}
	arpClient, err := arp.Dial(iface)
	if err != nil {
		return err
	}
	destMac, err := arpClient.Resolve(gatewayIP)
	if err != nil {
		return err
	}
	buf := gopacket.NewSerializeBuffer()
	serializeOpts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}
	ethLayer := &layers.Ethernet{
		SrcMAC:       iface.HardwareAddr,
		DstMAC:       destMac,
		EthernetType: layers.EthernetTypeIPv4,
	}
	ipLayer := &layers.IPv4{
		SrcIP:    sourceIP,
		DstIP:    serverIP,
		Protocol: layers.IPProtocolUDP,
		Version:  4,
		TTL:      32,
	}
	udpLayer := &layers.UDP{
		SrcPort: layers.UDPPort(c.port),
		DstPort: layers.UDPPort(serverPort),
	}
	udpLayer.SetNetworkLayerForChecksum(ipLayer)
	err = gopacket.SerializeLayers(buf, serializeOpts, ethLayer, ipLayer, udpLayer, gopacket.Payload([]byte(message)))
	if err != nil {
		return err
	}
	handle, err := pcap.OpenLive(iface.Name, 1024, false, pcap.BlockForever)
	if err != nil {
		log.Println(fmt.Errorf("error creating handle: %w", err))
	}
	if expectReply {
		go func() {
			reply, srcip, srcport, err := c.awaitReply(handle, serverIP, uint16(serverPort), 5*time.Second)
			if err != nil {
				log.Println(fmt.Errorf("error awaiting reply: %w", err))
			}
			c.handleReply(reply, srcip, srcport)
		}()
	}
	return handle.WritePacketData(buf.Bytes())
}

func (c *Client) awaitReply(handle *pcap.Handle, serverIP net.IP, serverPort uint16, timeout time.Duration) (reply string, srcIP net.IP, srcPort string, err error) {
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	packets := packetSource.Packets()
	for {
		select {
		case <-time.After(timeout):
			return "", nil, "", fmt.Errorf("reached timeout")
		case packet := <-packets:
			ip4Layer := packet.Layer(layers.LayerTypeIPv4)
			if ip4Layer == nil {
				break
			}
			var ip4 *layers.IPv4
			var ok bool
			if ip4, ok = ip4Layer.(*layers.IPv4); !ok {
				return "", nil, "", fmt.Errorf("ipv4 layer is not ipv4 layer")
			}
			if !ip4.SrcIP.Equal(serverIP) {
				break
			}
			udpLayer := packet.Layer(layers.LayerTypeUDP)
			if udpLayer == nil {
				break
			}
			var udp *layers.UDP
			if udp, ok = udpLayer.(*layers.UDP); !ok {
				return "", nil, "", fmt.Errorf("udp layer is not udp layer")
			}
			if udp.SrcPort != layers.UDPPort(serverPort) {
				fmt.Println(udp.SrcPort)
				break
			}
			return string(udp.Payload), ip4.SrcIP, udp.SrcPort.String(), nil
		}
	}
}

func (c *Client) handleReply(message string, srcIP net.IP, srcPort string) {
	args := utils.GetQuery(message)
	switch args[0] {
	// a <public_key> <ip>
	case constants.AnnounceQuery:
		if len(args) < 3 {
			log.Println(fmt.Errorf("wrong number of args: is %d but should be more than %d", len(args), 3))
		}
		ip := args[2]
		//TODO: check if from server
		publicKey, err := wgtypes.ParseKey(args[1])
		if err != nil {
			log.Println(fmt.Errorf("public key parsing failed: %w", err))
			return
		}
		ownPublicKey, err := c.manager.GetPublicKey()
		if err != nil {
			log.Println(fmt.Errorf("error getting own publickey: %w", err))
			return
		}
		if publicKey != *ownPublicKey {
			log.Println(fmt.Errorf("received reply has a different public key (%s) than our own (%s)", publicKey.String(), ownPublicKey.String()))
			return
		}
		log.Printf("received from %s:%s with content: %s %s %s", srcIP.String(), srcPort, constants.AnnounceQuery, publicKey, ip)
		return
	// g <peer>
	case constants.GetQuery:
		if len(args) < 2 {
			log.Println(fmt.Errorf("wrong number of args: is %d but should be more than %d", len(args), 2))
		}
		//TODO: check if from server
		reply := &connection.GetReply{}
		err := json.Unmarshal([]byte(args[1]), reply)
		if err != nil {
			log.Println(fmt.Errorf("unmarshalling reply failed: %w", err))
			return
		}
		publicKey, err := wgtypes.ParseKey(reply.PublicKey)
		if err != nil {
			log.Println(fmt.Errorf("reply public key parsing failed: %w", err))
			return
		}
		ownPublicKey, err := c.manager.GetPublicKey()
		if err != nil {
			log.Println(fmt.Errorf("error getting own publickey: %w", err))
			return
		}
		if publicKey == *ownPublicKey {
			log.Println(fmt.Errorf("peer has the same public key: %s", publicKey.String()))
			return
		}
		endpointAddr, err := net.ResolveUDPAddr("udp", reply.Endpoint)
		if err != nil {
			log.Println(fmt.Errorf("endpoint resolution failed: %w", err))
			return
		}
		cidr := reply.CIDR
		//TODO: check set our ip to something compatible
		keepalive := constants.DEFAULT_KEEPALIVE
		err = c.manager.SetPeer(publicKey, cidr, endpointAddr, &keepalive)
		if err != nil {
			log.Println(fmt.Errorf("setting config to peer failed: %w", err))
			return
		}
		log.Printf("peer set to: %s", reply)
		return

	}
}
