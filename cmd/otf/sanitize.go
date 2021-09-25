package main

import (
	"fmt"
	"net/url"
)

// Ensure hostname is in the format <host>:<port>
func sanitizeHostname(hostname string) (string, error) {
	u, err := url.ParseRequestURI(hostname)
	if err != nil || u.Host == "" {
		u, er := url.ParseRequestURI("https://" + hostname)
		if er != nil {
			return "", fmt.Errorf("could not parse hostname: %w", err)
		}
		return u.Host, nil
	}

	return u.Host, nil
}

// Ensure address is in format https://<host>:<port>
func sanitizeAddress(address string) (string, error) {
	u, err := url.ParseRequestURI(address)
	if err != nil || u.Host == "" {
		u, er := url.ParseRequestURI("https://" + address)
		if er != nil {
			return "", fmt.Errorf("could not parse hostname: %w", err)
		}
		return u.String(), nil
	}

	u.Scheme = "https"
	return u.String(), nil
}
