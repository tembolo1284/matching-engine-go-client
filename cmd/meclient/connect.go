// Full path: cmd/meclient/connect.go

package main

import (
	"fmt"
	"time"

	"github.com/tembolo1284/matching-engine-go-client/pkg/meclient"
)

// connectWithTransport connects using a specific transport and protocol.
func connectWithTransport(addr string, transport meclient.Transport, binary bool) (*meclient.Client, error) {
	cfg := meclient.DefaultConfig(addr)
	cfg.Transport = transport
	cfg.AutoReconnect = (transport == meclient.TransportTCP)

	if binary {
		cfg.Protocol = meclient.ProtocolBinary
	} else {
		cfg.Protocol = meclient.ProtocolCSV
	}

	transportStr := "TCP"
	if transport == meclient.TransportUDP {
		transportStr = "UDP"
	}
	protocolStr := "CSV"
	if binary {
		protocolStr = "binary"
	}

	fmt.Printf("Connecting to %s via %s (protocol: %s)...\n", addr, transportStr, protocolStr)

	client, err := meclient.New(cfg)
	if err != nil {
		return nil, err
	}

	if err := client.Connect(); err != nil {
		return nil, err
	}

	return client, nil
}

// connectWithFallback tries TCP first, falls back to UDP if TCP fails.
func connectWithFallback(addr string, binary bool) (*meclient.Client, error) {
	fmt.Printf("Connecting to %s via TCP...\n", addr)

	cfg := meclient.DefaultConfig(addr)
	cfg.Transport = meclient.TransportTCP
	cfg.ConnectTimeout = 2 * time.Second

	if binary {
		cfg.Protocol = meclient.ProtocolBinary
	} else {
		cfg.Protocol = meclient.ProtocolCSV
	}

	client, err := meclient.New(cfg)
	if err != nil {
		return nil, err
	}

	if err := client.Connect(); err != nil {
		fmt.Printf("TCP connection failed: %v\n", err)
		fmt.Printf("Falling back to UDP...\n")

		cfg.Transport = meclient.TransportUDP
		cfg.AutoReconnect = false

		client, err = meclient.New(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create UDP client: %w", err)
		}

		if err := client.Connect(); err != nil {
			return nil, fmt.Errorf("UDP connection also failed: %w", err)
		}

		protocolStr := "CSV"
		if binary {
			protocolStr = "binary"
		}
		fmt.Printf("Connected via UDP (protocol: %s)\n", protocolStr)
		return client, nil
	}

	protocolStr := "CSV"
	if binary {
		protocolStr = "binary"
	}
	fmt.Printf("Connected via TCP (protocol: %s)\n", protocolStr)
	return client, nil
}
