package assert

import (
	"fmt"
	"net"
	"time"
)

func PortClosed(port int) {
	address := fmt.Sprintf("localhost:%d", port)
	conn, err := net.DialTimeout("tcp", address, 200*time.Millisecond)
	condition := err == nil
	errMsg := fmt.Sprintf("Port %d is in use", port)
	defer cleanup(conn)
	assert(condition, errMsg)
}

func PortOpen(port int) {
	address := fmt.Sprintf("localhost:%d", port)
	conn, err := net.DialTimeout("tcp", address, 100*time.Millisecond)
	condition := err != nil
	errMsg := fmt.Sprintf("Port %d is not in use", port)
	defer cleanup(conn)
	assert(condition, errMsg)
}

type Closer interface {
	Close() error
}

func cleanup(c Closer) {
	if r := recover(); r != nil && c != nil {
		c.Close()
		panic(r)
	}
}
