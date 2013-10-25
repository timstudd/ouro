package main

import (
	"./nat"
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
	peerConnections	map[string]*PeerConn
}

type PeerConn struct {
	sideband	*ShimConn
	initiator	bool
	udpConn		net.Conn
	cryptConn	*EncryptedConnection
	ignorePkts	bool
}

func NewUser() *User {
	user := &User{peerConnections: make(map[string]*PeerConn)}
	return user
}

// func (self *User) Authenticate() error {
// 	var userId string

// 	userId = ""
// 	if self.Id != "" {
// 		userId = self.Id
// 	}

// 	env := Envelope{Auth: &Auth{UUID:userId}}
// 	err := self.encoder.Encode(&env)
// 	return err
// }

// func (self *User) Connect(host string) error {
// 	err := self.getTrackerConnection(host)
// 	if err != nil {
// 		return err
// 	}

// 	self.getEncodeDecode()
// 	self.Authenticate()

// 	return nil
// }

// func (self *User) getTrackerConnection(host string) (error) {
// 	addr, err := net.ResolveTCPAddr("tcp", host)
// 	if err != nil {
// 		return err
// 	}

// 	self.conn, err = net.DialTCP("tcp", nil, addr)
// 	if err != nil {
// 		return err
// 	}

// 	self.cryptConn = &EncryptedConnection{Destination: self.conn}

// 	return nil
// }

// func (self *User) getEncodeDecode() error {
// 	if self.conn != nil {
// 		self.encoder = gob.NewEncoder(self.conn)
// 		self.decoder = gob.NewDecoder(self.conn)
// 		return nil
// 	}
// 	return errors.New("Invalid connection")
// }

func (self *User) HandlePcSignal(signal PcSignal) {
	pc, ok := self.peerConnections[signal.From]
	if !ok {
		pc = self.MakePeerConn(signal.From, false)
	}
	pc.sideband.readChan <- signal.Payload
}

func (self *User) MakePeerConn(peerId string, initiator bool) *PeerConn {
	pc := &PeerConn{
		sideband:   newShimConn(self.encoder, peerId),
		initiator:  initiator,
		udpConn:    nil,
		ignorePkts: true,
	}
	self.peerConnections[peerId] = pc

	go func() {
		var err error
		pc.udpConn, err = nat.Connect(pc.sideband, pc.initiator)
		pc.cryptConn = &EncryptedConnection{Destination: pc.udpConn}
		if err != nil {
			log.Println("err doing nat conn", err)
			// TODO REMOVE FROM MAP
		} else {
			go func() {
				pc.ignorePkts = false
				pc.cryptConn.Write([]byte("Established"))
			}()
			handleRemoteUdp(pc)
		}
	}()

	return pc
}

func (self * User) closePeerConnections() {
	for _, v := range self.peerConnections {
		closeRemoteUdp(v)
	}
	self.peerConnections = make(map[string]*PeerConn)
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

func closeRemoteUdp(pc *PeerConn) (error) {
	err := pc.udpConn.Close()
	if err != nil {
		panic(err)
	}
	return err
}