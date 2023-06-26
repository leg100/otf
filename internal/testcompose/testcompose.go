// Package testcompose provides interaction with a docker compose stack of
// services for testing purposes.
package testcompose

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

const (
	Postgres Service = "postgres"
	Squid    Service = "squid"
	PubSub   Service = "pubsub"
)

type Service string

var ports = map[Service]int{
	Postgres: 5432,
	Squid:    3128,
	PubSub:   8085,
}

func Up() error {
	// --wait implies -d, which detaches the containers
	args := []string{"compose", "up", "--wait", "--wait-timeout", "60"}
	args = append(args, string(Postgres), string(Squid), string(PubSub))
	cmd := exec.Command("docker", args...)

	var buferr bytes.Buffer
	cmd.Stderr = &buferr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%v failed: %s: %w", args, buferr.String(), err)
	}
	return nil
}

func GetHost(svc Service) (string, error) {
	port, ok := ports[svc]
	if !ok {
		return "", fmt.Errorf("service not found: %s", svc)
	}
	args := []string{"docker", "compose", "port", string(svc), strconv.Itoa(port)}
	cmd := exec.Command(args[0], args[1:]...)

	var bufout, buferr bytes.Buffer
	cmd.Stdout = &bufout
	cmd.Stderr = &buferr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("getting port info for %s: %s: %w", svc, buferr.String(), err)
	}

	parts := strings.Split(bufout.String(), ":")
	if len(parts) != 2 {
		return "", fmt.Errorf("unexpected output from %v: %v", args, parts)
	}
	return fmt.Sprintf("localhost:%s", parts[1]), nil
}
