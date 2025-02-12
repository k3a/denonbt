//go:build linux

package main

import (
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

type RFCommSocket struct {
	fd        int
	connected bool
}

func str2ba(addr string) [6]byte {
	a := strings.Split(addr, ":")
	var b [6]byte
	for i, tmp := range a {
		u, _ := strconv.ParseUint(tmp, 16, 8)
		b[len(b)-1-i] = byte(u)
	}
	return b
}

func (sock *RFCommSocket) IsConnected() bool {
	return sock.connected
}

func (sock *RFCommSocket) Connect(hwaddr string, channel uint8) error {
	var err error

	// close old socket
	if sock.fd > 0 {
		sock.Close()
		time.Sleep(3 * time.Second)
	}

	// create socket
	sock.fd, err = unix.Socket(syscall.AF_BLUETOOTH, syscall.SOCK_STREAM, unix.BTPROTO_RFCOMM)
	if err != nil {
		return err
	}

	addr := &unix.SockaddrRFCOMM{Addr: str2ba(hwaddr), Channel: channel}

	err = unix.Connect(sock.fd, addr)
	sock.connected = err == nil
	return err
}

func (sock *RFCommSocket) Read(data []byte) (int, error) {
	return unix.Read(sock.fd, data)
}

func (sock *RFCommSocket) Write(data []byte) (int, error) {
	return unix.Write(sock.fd, data)
}

func (sock *RFCommSocket) SetNonBlocking() error {
	_, err := unix.FcntlInt(uintptr(sock.fd), unix.F_SETFL, unix.O_NONBLOCK)
	return err
}

func (sock *RFCommSocket) Close() error {
	err := unix.Close(sock.fd)
	sock.fd = 0
	sock.connected = false
	return err
}
