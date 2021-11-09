package util

import (
	"fmt"
	"net"
)

func NewAutoAddressNetListener() (
	ip string, port int, listener net.Listener,
) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}

	ip = string(listener.Addr().(*net.TCPAddr).IP)
	port = listener.Addr().(*net.TCPAddr).Port
	return
}

func NewNetListener(
	host string, port int,
) (listener net.Listener) {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		panic(err)
	}

	return
}
