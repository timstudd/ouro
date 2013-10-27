package main

import (
	"net"
	"encoding/gob"
	"log"
	"fmt"
	"time"
)

type User struct {
	Id				string
	Name			string
	trackers		[]*Tracker
	conn			net.Conn
	cryptConn		*EncryptedConnection
	encoder			*gob.Encoder
	decoder			*gob.Decoder
}

type PeerConn struct {
	sideband	*ShimConn
	initiator	bool
	udpConn		net.Conn
	cryptConn	*EncryptedConnection
	ignorePkts	bool
}

func NewUser() *User {
	user := &User{}
	return user
}

func handleRemoteUdp(pc *PeerConn) {
	for {
		data := make([]byte, 65535)
		_, err := pc.cryptConn.Read(data)

		if err != nil {
			log.Println("Lost peer connection")
			return
		} else if !pc.ignorePkts {
			log.Println("Received:", string(data))

			time.Sleep(1 * time.Second)
			send := fmt.Sprintf("Hi %s", time.Now().String())
			log.Println("Sent:", send)
			pc.cryptConn.Write([]byte(send))
		}
	}
}