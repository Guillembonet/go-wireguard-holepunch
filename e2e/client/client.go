package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"

	"github.com/guillembonet/go-wireguard-udpholepunch/e2e/common/di"
	"github.com/guillembonet/go-wireguard-udpholepunch/e2e/common/params"
	"github.com/guillembonet/go-wireguard-udpholepunch/e2e/common/server"
	"github.com/guillembonet/go-wireguard-udpholepunch/e2e/common/server/endpoints"
)

func main() {
	container := &di.Container{}
	defer container.Cleanup()

	var gparams params.Generic
	gparams.Init()

	var cparams params.Client
	cparams.Init()

	flag.Parse()

	// use router
	gateway := os.Getenv("GATEWAY")
	if gateway != "" {
		err := exec.Command("ip", "route", "del", "default").Run()
		if err != nil {
			panic(err)
		}
		err = exec.Command("ip", "route", "add", "default", "via", gateway).Run()
		if err != nil {
			panic(err)
		}
	}

	errChan := make(chan error)

	client, err := container.ConstructClient(gparams, cparams)
	if err != nil {
		panic(err)
	}

	ae := endpoints.NewAnnounceEndpoint(client, *cparams.ServerIP, *cparams.ServerPort)
	server := server.NewServer(ae)
	if err != nil {
		panic(err)
	}

	go func() {
		errChan <- server.Serve()
	}()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)

	go func() {
		<-sigchan
		errChan <- fmt.Errorf("received an interrupt signal")
	}()

	err = <-errChan
	log.Println(err)
}
