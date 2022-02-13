package e2e

import (
	"fmt"
	"time"

	"github.com/magefile/mage/sh"
)

// E2E runs the e2e tests
func E2E() error {
	_, err := StartCompose()
	if err != nil {
		return fmt.Errorf("could not deploy containers")
	}
	defer TeardownCompose()

	time.Sleep(2 * time.Second)

	err = sh.RunV("go",
		"test",
		"-v",
		"-tags=e2e",
		"./e2e",
	)
	return err
}

func StartCompose() (string, error) {
	return sh.Output("docker-compose", "-f", "./e2e/docker-compose.e2e.yml", "up", "--build", "-d")
}

func TeardownCompose() (string, error) {
	return sh.Output("docker-compose", "-f", "./e2e/docker-compose.e2e.yml", "down")
}
