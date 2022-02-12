package e2e

import (
	"fmt"

	"github.com/magefile/mage/sh"
)

const (
	// DefaultConsumerBalance is used as default balance for hermes mock
	DEFAULT_CONSUMER_BALANCE uint64 = 10
	JWT_SECRET_E2E                  = "token"
)

// E2E runs the e2e tests
func E2E() error {
	_, err := StartCompose()
	if err != nil {
		return fmt.Errorf("could not deploy containers")
	}
	defer TeardownCompose()

	err = sh.RunV("go",
		"test",
		"-v",
		"-tags=e2e",
		"./e2e",
	)
	return err
}

func StartCompose() (string, error) {
	return sh.Output("docker-compose", "-f", "docker-compose.e2e.yml", "up", "-d")
}

func TeardownCompose() (string, error) {
	return sh.Output("docker-compose", "-f", "docker-compose.e2e.yml", "down")
}
