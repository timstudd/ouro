package main

import (
	"./nat"
	// "./utils"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"time"
)

var peerConnections map[string]*PeerConn

type PeerConn struct {
	sideband   *ShimConn
	initiator  bool
	udpConn    net.Conn
	ignorePkts bool
}

func runClient(host string) error {
	var err error
	var self User

	peerConnections = make(map[string]*PeerConn)

	conn, err := getTrackerConnection(host)
	if err != nil {
		return err
	}
	encoder, decoder := getEncodeDecode(conn)

	err = loginTracker(encoder, self)
	if err != nil {
		return err
	}

	for {
		var env Envelope
		err := decoder.Decode(&env)
		if err != nil {
			// Close and reset peerConnections
			closePeerConnections(peerConnections)
			peerConnections = make(map[string]*PeerConn)

			time.Sleep(1 * time.Second)
			conn, err := getTrackerConnection(host)
			if err == nil {
				encoder, decoder = getEncodeDecode(conn)
				_ = loginTracker(encoder, self)
			}
		}

		if env.Auth != nil && env.Auth.UUID != "" {
			self.Id = env.Auth.UUID
			log.Println("Got User Id:", self.Id)
		}

		if env.PcSignal != nil {
			HandlePcSignal(encoder, *env.PcSignal)
		}

		if env.UserList != nil {
			// This is the list of existing users.
			// Let's try to establish a PeerConn to each
			for _, u := range env.UserList {
				MakePeerConn(encoder, u.Id, true)
			}
		}
	}

	return nil
}

func getEncodeDecode(conn net.Conn) (*gob.Encoder, *gob.Decoder) {
	encoder := gob.NewEncoder(conn)
	decoder := gob.NewDecoder(conn)

	return encoder, decoder
}

func loginTracker(encoder *gob.Encoder, self User) error {
	var userId string

	userId = ""
	if self.Id != "" {
		userId = self.Id
	}

	env := Envelope{Auth: &Auth{UUID:userId}}
	err := encoder.Encode(&env)
	return err
}

func getTrackerConnection(host string) (net.Conn, error) {
	addr, err := net.ResolveTCPAddr("tcp", host)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func HandlePcSignal(encoder *gob.Encoder, signal PcSignal) {
	pc, ok := peerConnections[signal.From]
	if !ok {
		pc = MakePeerConn(encoder, signal.From, false)
	}
	pc.sideband.readChan <- signal.Payload
}

func MakePeerConn(encoder *gob.Encoder, peerId string, initiator bool) *PeerConn {
	pc := &PeerConn{
		sideband:   newShimConn(encoder, peerId),
		initiator:  initiator,
		udpConn:    nil,
		ignorePkts: true,
	}
	peerConnections[peerId] = pc

	go func() {
		var err error
		pc.udpConn, err = nat.Connect(pc.sideband, pc.initiator)
		if err != nil {
			log.Println("err doing nat conn", err)
			// TODO REMOVE FROM MAP
		} else {
			go func() {
				pc.ignorePkts = false
				pc.udpConn.Write([]byte("Established"))
			}()
			handleRemoteUdp(pc)
		}
	}()

	return pc
}

func closePeerConnections(connections map[string]*PeerConn) {
	for _, v := range connections {
		closeRemoteUdp(v)
	}
}

func handleRemoteUdp(pc *PeerConn) {
	data := make([]byte, 65535)
	for {
		_, err := pc.udpConn.Read(data)

		if err != nil {
			log.Println("Lost peer connection")
			return
		} else if !pc.ignorePkts {
			log.Println("Received:", string(data))

			time.Sleep(1 * time.Second)
			send := fmt.Sprintf("Hi %s", time.Now().String())
			pc.udpConn.Write([]byte(send))
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