//go:build darwin

package main

import (
	"errors"
)

type RFCommSocket struct {
}

var ErrNotImplemented = errors.New("rfcommsock: not implemented")

func (sock *RFCommSocket) IsConnected() bool {
	return false
}

func (sock *RFCommSocket) Connect(hwaddr string, channel uint8) error {
	return ErrNotImplemented
}

func (sock *RFCommSocket) Read(data []byte) (int, error) {
	return 0, ErrNotImplemented
}

func (sock *RFCommSocket) Write(data []byte) (int, error) {
	return 0, ErrNotImplemented
}

func (sock *RFCommSocket) SetNonBlocking() error {
	return ErrNotImplemented
}

func (sock *RFCommSocket) Close() error {
	return ErrNotImplemented
}
