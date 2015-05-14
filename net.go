package dht

import (
	"net"
	"github.com/armon/go-metrics"
)

func (dht *Dht) sendMsg(destination net.Addr, msg []byte) error {
	_, err := dht.listener.WriteTo(msg, destination)
	if err == nil {
		metrics.IncrCounter([]string{"kad-dht", "udp", "sent"}, float32(len(msg)))
	}
	return err
}

func (dht *Dht) udpListen() {
	var n int
	var addr net.Addr
	var err error
	var lastPacket time.Time
	for {
		// Do a check for potentially blocking operations
		if !lastPacket.IsZero() && time.Now().Sub(lastPacket) > blockingWarning {
			diff := time.Now().Sub(lastPacket)
			m.logger.Printf(
				"[DEBUG] memberlist: Potential blocking operation. Last command took %v",
				diff)
		}

		// Create a new buffer
		// TODO: Use Sync.Pool eventually
		// GY: basically using pool enables you to reuse buffer => relieving G pressure
		buf := make([]byte, udpBufSize)

		// Read a packet
		n, addr, err = m.udpListener.ReadFrom(buf)
		if err != nil {
			if m.shutdown {
				break
			}
			m.logger.Printf("[ERR] memberlist: Error reading UDP packet: %s", err)
			continue
		}

		// Check the length
		if n < 1 {
			m.logger.Printf("[ERR] memberlist: UDP packet too short (%d bytes). From: %s",
				len(buf), addr)
			continue
		}

		// Capture the current time
		lastPacket = time.Now()

		// Ingest this packet
		metrics.IncrCounter([]string{"memberlist", "udp", "received"}, float32(n))
		m.ingestPacket(buf[:n], addr) // GY: where the main action happens => eventually gets handed off to handler
	}
}

func (dht *Dht) udpHandler() {
	for {
		select {
		case msg := <-m.handoff:
			msgType := msg.msgType
			buf := msg.buf
			from := msg.from

			switch msgType {
			case suspectMsg:
				m.handleSuspect(buf, from)
			case aliveMsg:
				m.handleAlive(buf, from)
			case deadMsg:
				m.handleDead(buf, from)
			case userMsg:
				m.handleUser(buf, from)
			default:
				m.logger.Printf("[ERR] memberlist: UDP msg type (%d) not supported. From: %s (handler)", msgType, from)
			}

		case <-m.shutdownCh:
			return
		}
	}
}

func setUDPBufSize(conn *net.UDPConn) {
	size := udpBufSize
	for {
		if err := conn.SetReadBuffer(size); err == nil {
			break
		}
		size = size / 2
	}
}