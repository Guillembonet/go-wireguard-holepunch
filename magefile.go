//go:build mage

package main

import (
	"github.com/guillembonet/go-wireguard-holepunch/e2e"
)

// E2E runs the e2e test suite
func E2E() error {
	return e2e.E2E()
}
