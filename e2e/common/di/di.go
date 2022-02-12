package di

import (
	"strconv"
	"sync"

	"github.com/guillembonet/go-wireguard-holepunch/communication"
	"github.com/guillembonet/go-wireguard-holepunch/communication/client"
	"github.com/guillembonet/go-wireguard-holepunch/communication/server"
	"github.com/guillembonet/go-wireguard-holepunch/connection"
	"github.com/guillembonet/go-wireguard-holepunch/e2e/common/params"
	"github.com/guillembonet/go-wireguard-holepunch/storage"
)

// Container represents our dependency container
type Container struct {
	cleanup []func()
	lock    sync.Mutex
}

// Cleanup performs the cleanup required
func (c *Container) Cleanup() {
	c.lock.Lock()
	defer c.lock.Unlock()
	for i := len(c.cleanup) - 1; i >= 0; i-- {
		c.cleanup[i]()
	}
}

// // ConstructServer creates a server for us
func (c *Container) ConstructServer(gparams params.Generic) (*server.Server, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	sock, err := communication.CreateUDPSocket(":" + strconv.Itoa(*gparams.Port))
	if err != nil {
		return nil, err
	}
	c.cleanup = append(c.cleanup, func() { sock.Close() })
	storage := storage.NewStorage()
	server := server.NewServer(sock, storage)
	return server, nil
}

// ConstructServer creates a server for us
func (c *Container) ConstructClient(gparams params.Generic, cparams params.Client) (*client.Client, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	manager, err := connection.NewManager(*cparams.WireguardInterfaceName, *gparams.Port)
	if err != nil {
		return nil, err
	}
	c.cleanup = append(c.cleanup, func() { manager.Cleanup() })

	client := client.NewClient(*gparams.Port, manager)
	return client, nil
}
