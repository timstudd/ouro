package main

import (
	"./nat"
	"net"
	"github.com/lunny/xorm"
	"encoding/gob"
	"log"
	"time"
	"errors"
)

type Tracker struct {
	Id				int64
	Host			string
	UUID			string
	Active			bool
	conn			net.Conn			`xorm:"-"`
	encoder			*gob.Encoder		`xorm:"-"`
	decoder 		*gob.Decoder		`xorm:"-"`
	peerConnections	map[string]*PeerConn `xorm:"-"`
}

func GetTrackers(orm *xorm.Engine) ([]Tracker, error) {
	trackers := make([]Tracker, 0)
	err := orm.Find(&trackers)
	if err != nil {
		return nil, err
	}	

	return trackers, err
}

func (self *Tracker) Connect() (error) {
	addr, err := net.ResolveTCPAddr("tcp", self.Host)
	if err != nil {
		return err
	}

	self.conn, err = net.DialTCP("tcp", nil, addr)
	if err != nil {
		return err
	}

	if self.conn != nil {
		self.encoder = gob.NewEncoder(self.conn)
		self.decoder = gob.NewDecoder(self.conn)
	} else {
		return errors.New("Invalid connection")
	}

	self.peerConnections = make(map[string]*PeerConn)

	log.Println("Connected to ", self.Host)

	return nil
}

func (self *Tracker) Authenticate() error {
	userId := ""
	if self.UUID != "" {
		userId = self.UUID
	}

	env := &Envelope{Auth: &Auth{UUID:userId}}
	err := self.encoder.Encode(env)

	log.Println("Sent authenticate UUID:", userId)

	return err
}

func (self *Tracker) MakePeerConn(peerId string, initiator bool) *PeerConn {
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
		if err != nil {
			log.Println("err doing nat conn", err)
			// TODO REMOVE FROM MAP
		} else {
			pc.cryptConn = &EncryptedConnection{Destination: pc.udpConn}
			go func() {
				pc.ignorePkts = false
				pc.cryptConn.Write([]byte("Established"))
			}()
			handleRemoteUdp(pc)
		}
	}()

	return pc
}

func (self *Tracker) HandlePcSignal(signal PcSignal) {
	pc, ok := self.peerConnections[signal.From]
	if !ok {
		pc = self.MakePeerConn(signal.From, false)
	}
	pc.sideband.readChan <- signal.Payload
}

func (self * Tracker) closePeerConnections() {
	for _, v := range self.peerConnections {
		closeRemoteUdp(v)
	}
	// Set peer connections to empty
	self.peerConnections = make(map[string]*PeerConn)
}

func closeRemoteUdp(pc *PeerConn) (error) {
	err := pc.udpConn.Close()
	if err != nil {
		panic(err)
	}
	return err
}

func (tracker *Tracker) Listen(orm *xorm.Engine, user *User) {
	for {
		var env Envelope
		err := tracker.decoder.Decode(&env)

		log.Println("Got message")

		if err != nil {
			// Close and reset peerConnections
			tracker.closePeerConnections()

			// Try to reconnect
			time.Sleep(1 * time.Second)
			err = tracker.Connect()
			if err != nil {
				go tracker.Listen(orm, user)
				err = tracker.Authenticate()
				if err != nil {
					log.Println("error authenticating")
				}
			}
		}

		if env.Auth != nil && env.Auth.UUID != "" {
			tracker.UUID = env.Auth.UUID
			_, err = orm.Id(tracker.Id).Update(tracker)
			if err != nil {
				log.Println("Error updating:", err)
			}
			log.Println("Got User Id:", tracker.UUID)
		}

		if env.PcSignal != nil {
			log.Println("Got pc sig")
			tracker.HandlePcSignal(*env.PcSignal)
		}

		if env.UserList != nil {
			log.Println("Got users")
			// Received list of users - try to establish a PeerConn to each
			for _, u := range env.UserList {
				log.Println("Making PC:", u.UUID)
				tracker.MakePeerConn(u.UUID, true)
			}
		}
	}
}