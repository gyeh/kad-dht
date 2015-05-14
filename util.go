package dht
import (
	"crypto/rand"
	"fmt"
)

// Create a 160-bit random ID
func createRandomId() []byte {
	b := make([]byte, 20)
	_, err := rand.Read(b)
	if err != nil {
		panic(fmt.Sprintf("Unable to create id. Error %v", err))
	}
	return b
}
