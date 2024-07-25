package mocks

import (
	"io"
)

type MockChannel struct {
	io.Reader
	io.Writer
	io.Closer
}

func (c *MockChannel) SendRequest(name string, wantReply bool, payload []byte) (bool, error) {
	return true, nil
}

func (c *MockChannel) Stderr() io.ReadWriter {
	return c
}

func (c *MockChannel) CloseWrite() error {
	return nil
}

func (c *MockChannel) Read(data []byte) (int, error) {
	return len(data), nil
}

func (c *MockChannel) Write(data []byte) (int, error) {
	return len(data), nil
}

func (c *MockChannel) Close() error {
	return nil
}
