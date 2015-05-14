package dht

import (
	"net"
	"sync"
	"log"
	"fmt"
	"os"
)

type Dht struct {
	config *Config
	nodeId []byte

	listener *net.UDPConn

	isClosed       bool
	closedCh     chan struct{}
	closeLock sync.Mutex

	logger *log.Logger
}

func newDht(config *Config) (*Dht, error) {
	addr := &net.UDPAddr{IP: net.ParseIP(config.BindAddr), Port: config.BindPort}
	listener, err := net.ListenUDP("udp", addr)
	if err != nil {
		listener.Close()
		return nil, fmt.Errorf("Unable to start listener. Error: %s", err)
	}

	setUDPBufSize(listener)

	if config.LogOut == nil {
		config.LogOut = os.Stderr
	}
	logger := log.New(config.LogOut, "", log.LstdFlags)

	dht := &Dht {
		config: config,
		nodeId: createRandomId(),
		listener:    listener,
		closedCh:     make(chan struct{}),
		logger: logger,
	}
	go dht.udpListen()
	go dht.udpHandler()

	return dht, nil
}

func (dht *Dht) Close() error {
	dht.closeLock.Lock()
	defer dht.closeLock.Unlock()

	if dht.isClosed {
		return nil
	}

	dht.isClosed = true
	close(dht.closedCh)
	dht.listener.Close()
	return nil
}