package dht
import "io"

const (
	udpBufSize = 1024 * 1024 * 2
)

type Config struct {
	BucketSize int

	// UDP addr and port for incoming traffic.
	BindAddr string
	BindPort int

	// Writer where logs are sent. Stderr by default.
	LogOut io.Writer
}