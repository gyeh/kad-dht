package dht
import (
	"net"
	"github.com/armon/go-metrics"
	"time"
)

// ping request sent directly to node
type ping struct {
	SeqNo uint32

	// Node is sent so the target can verify they are
	// the intended recipient. This is to protect again an agent
	// restart with a new name.
	Node string
}

type pongHandler struct {
	handler func()
	timer   *time.Timer
}

func (dht *Dht) Ping(net.IP, port string) {
	defer metrics.MeasureSince([]string{"kad-dht", "ping"}, time.Now())
}

func (dht *Dht) setAckChannel(seqNo uint32, ch chan bool, timeout time.Duration) {
	handler := func() {
		select {
		case ch <- true:
		default:
		}
	}

	ah := &pongHandler{handler, nil}
	dht.ackLock.Lock()
	dht.pongHandlers[seqNo] = ah
	dht.ackLock.Unlock()

	// Setup a reaping routing
	ah.timer = time.AfterFunc(timeout, func() {
		dht.ackLock.Lock()
		delete(dht.pongHandlers, seqNo)
		dht.ackLock.Unlock()
		select {
		case ch <- false:
		default:
		}
	})
}

func (dht *Dht) invokepongHandler(seqNo uint32) {
	dht.ackLock.Lock()
	ah, ok := dht.pongHandlers[seqNo]
	delete(dht.pongHandlers, seqNo)
	dht.ackLock.Unlock()
	if !ok {
		return
	}
	ah.timer.Stop()
	ah.handler()
}
