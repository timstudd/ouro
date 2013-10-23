package main

import (
	"encoding/gob"
	"net"
	"time"
)

type ShimConn struct {
	encoder  *gob.Encoder
	to       string
	readChan chan []byte
}

func newShimConn(encoder *gob.Encoder, to string) *ShimConn {
	return &ShimConn{encoder: encoder, to: to, readChan: make(chan []byte)}
}

func (sc *ShimConn) Write(bytes []byte) (int, error) {
	env := &Envelope{
		PcSignal: &PcSignal{
			To:      sc.to,
			Payload: bytes,
		},
	}
	err := sc.encoder.Encode(env)
	if err != nil {
		return 0, err
	}
	return len(bytes), nil
}

func (sc *ShimConn) Read(bytes []byte) (int, error) {
	tmp := <-sc.readChan
	n := copy(bytes, tmp)
	return n, nil
}

func (sc *ShimConn) LocalAddr() net.Addr                { return nil }
func (sc *ShimConn) Close() error                       { return nil }
func (sc *ShimConn) RemoteAddr() net.Addr               { return nil }
func (sc *ShimConn) SetDeadline(t time.Time) error      { return nil }
func (sc *ShimConn) SetReadDeadline(t time.Time) error  { return nil }
func (sc *ShimConn) SetWriteDeadline(t time.Time) error { return nil }