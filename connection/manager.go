package connection

import (
	"fmt"
	"net"
	"time"

	"github.com/guillembonet/go-wireguard-holepunch/utils"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// Manager is in charge of managing the wireguard connections
type Manager struct {
	iface     string
	client    *wgctrl.Client
	publicKey *wgtypes.Key
}

func NewManager(iface string, port int) (*Manager, error) {
	wgClient, err := wgctrl.New()
	if err != nil {
		return nil, err
	}
	manager := &Manager{
		iface:  iface,
		client: wgClient,
	}
	err = manager.initialize(port)
	if err != nil {
		return nil, err
	}
	return manager, nil
}

// Initializes the wireguard client with a new key and creates a device for it
func (m *Manager) initialize(port int) error {
	key, err := wgtypes.GenerateKey()
	if err != nil {
		return err
	}
	m.publicKey = &key
	err = m.createDevice()
	if err != nil {
		return err
	}
	config := wgtypes.Config{
		PrivateKey: &key,
		ListenPort: &port,
		Peers:      []wgtypes.PeerConfig{},
	}

	return m.client.ConfigureDevice(m.iface, config)
}

func (m *Manager) createDevice() error {
	if d, err := m.client.Device(m.iface); err != nil || d.Name != m.iface {
		err := utils.SudoExec("ip", "link", "add", "dev", m.iface, "type", "wireguard")
		if err != nil {
			fmt.Println(err)
			return err
		}
		err = utils.SudoExec("ip", "link", "set", "dev", m.iface, "up")
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("device with interface name: %s already exists", m.iface)
}

// SetPeer sets the wireguard peer
func (m *Manager) SetPeer(publicKey wgtypes.Key, cidr string, endpoint *net.UDPAddr, keepalive *time.Duration) error {
	device, err := m.client.Device(m.iface)
	if err != nil {
		return err
	}
	_, peerIps, err := net.ParseCIDR(cidr)
	if err != nil {
		return err
	}
	peer := wgtypes.PeerConfig{PublicKey: publicKey, PersistentKeepaliveInterval: keepalive, AllowedIPs: []net.IPNet{*peerIps}, Endpoint: endpoint}
	config := wgtypes.Config{
		PrivateKey:   &device.PrivateKey,
		ListenPort:   &device.ListenPort,
		Peers:        []wgtypes.PeerConfig{peer},
		ReplacePeers: true,
	}
	return m.client.ConfigureDevice(m.iface, config)
}

func (m *Manager) destroyDevice() error {
	return utils.SudoExec("ip", "link", "del", "dev", m.iface)
}

func (m *Manager) Cleanup() error {
	err := m.destroyDevice()
	if err != nil {
		return err
	}
	return m.client.Close()
}

func (m *Manager) GetInterfaceIP() (net.Addr, error) {
	iface, err := net.InterfaceByName(m.iface)
	if err != nil {
		return nil, err
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return nil, err
	}
	if len(addrs) != 1 {
		return nil, fmt.Errorf("interface %s doesn't have one ip address", m.iface)
	}
	ipv4Addr := addrs[0]
	if ipv4Addr == nil {
		return nil, fmt.Errorf("interface %s has a null ip", m.iface)
	}
	return ipv4Addr, nil
}

func (m *Manager) GetPublicKey() (*wgtypes.Key, error) {
	device, err := m.client.Device(m.iface)
	if err != nil {
		return nil, err
	}
	return &device.PublicKey, nil
}

func (m *Manager) GetPeers(publicKey wgtypes.Key) ([]wgtypes.Peer, error) {
	device, err := m.client.Device(m.iface)
	if err != nil {
		return nil, err
	}
	return device.Peers, nil
}

func (m *Manager) SetInterfaceIP(ip net.IP) error {
	if ip == nil {
		return fmt.Errorf("invalid ip")
	}
	if err := utils.SudoExec("ip", "address", "flush", "dev", m.iface); err != nil {
		return err
	}
	return utils.SudoExec("ip", "address", "replace", "dev", m.iface, ip.String())
}
