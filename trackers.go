package main

import (
	"net"
	"github.com/lunny/xorm"
	"encoding/gob"
	"log"
	"time"
	"errors"
)

type Tracker struct {
	Id		int64
	Host	string
	UUID	string
	Active	bool
	conn	net.Conn		`xorm:"-"`
	encoder	*gob.Encoder	`xorm:"-"`
	decoder *gob.Decoder	`xorm:"-"`
}

func GetTrackers(orm *xorm.Engine) ([]Tracker, error) {
	trackers := make([]Tracker, 0)
	err := orm.Find(&trackers)
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

	return err
}

func (self *Tracker) Listen() {
	for {
		var env Envelope
		err := self.decoder.Decode(&env)
		if err != nil {
			// Close and reset peerConnections
			// self.closePeerConnections()

			// Try to reconnect
			time.Sleep(1 * time.Second)
			self.Connect()
		}

		if env.Auth != nil && env.Auth.UUID != "" {
			// self.Id = env.Auth.UUID
			log.Println("Got User Id:", self.Id)
		}

		if env.PcSignal != nil {
			log.Println("Got pc sig")
			// self.HandlePcSignal(*env.PcSignal)
		}

		if env.UserList != nil {
			log.Println("Got users")
			// Received list of users - try to establish a PeerConn to each
			// for _, u := range env.UserList {
			// 	self.MakePeerConn(u.Id, true)
			// }
		}
	}
}