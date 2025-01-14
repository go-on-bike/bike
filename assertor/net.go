package assertor

import (
	"fmt"
	"net"
	"time"
)

func PortClosed(port int, msg string) {
	address := fmt.Sprintf("localhost:%d", port)
	conn, err := net.DialTimeout("tcp", address, 200*time.Millisecond)
	condition := err == nil
	errMsg := fmt.Sprintf("%s: Port %d is in use", msg, port)
	defer cleanup(conn)
	assert(condition, errMsg)
}

func PortOpen(port int, msg string) {
	address := fmt.Sprintf("localhost:%d", port)
	conn, err := net.DialTimeout("tcp", address, 100*time.Millisecond)
	condition := err != nil
    errMsg := fmt.Sprintf("%s: Port %d is not in use", msg, port)
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
