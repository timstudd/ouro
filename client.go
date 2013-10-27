package main

import (
	"log"
	"time"
)

func runClient(host string) error {
	// self := NewUser()

	engine, err := GetEngine()
	if err != nil {
		return err
	}
	// engine.ShowSQL = true

	err = engine.CreateTables(&Tracker{})
	if err != nil {
		return err
	}

	// Add default tracker if not exists
	defaultHost := &Tracker{Host: "192.168.1.64:9000", Active: true}
	has, err := engine.Get(defaultHost)
	if err != nil {
		return err
	}
	if !has {
		_, err = engine.Insert(defaultHost)
		if err != nil {
			return err
		}
	}

	trackers, err := GetTrackers(engine)
	if err != nil {
		return err
	}

	for _, tracker := range trackers {
		err = tracker.Connect()
		if err != nil {
			log.Println("Could not connect to ", tracker.Host)
		} else {
			go tracker.Listen(engine)
			err = tracker.Authenticate()
			if err != nil {
				return err
			}
		}
	}

	log.Println("Trackers:", len(trackers))

	for {
		time.Sleep(1 * time.Second)
	}

	// err = self.Connect(host)
	// if err != nil {
	// 	return err
	// }

	// for {
	// 	var env Envelope
	// 	err := self.decoder.Decode(&env)
	// 	if err != nil {
	// 		// Close and reset peerConnections
	// 		// self.closePeerConnections()

	// 		// Try to reconnect
	// 		time.Sleep(1 * time.Second)
	// 		self.Connect(host)
	// 	}

	// 	if env.Auth != nil && env.Auth.UUID != "" {
	// 		self.Id = env.Auth.UUID
	// 		log.Println("Got User Id:", self.Id)
	// 	}

	// 	if env.PcSignal != nil {
	// 		self.HandlePcSignal(*env.PcSignal)
	// 	}

	// 	if env.UserList != nil {
	// 		// Received list of users - try to establish a PeerConn to each
	// 		for _, u := range env.UserList {
	// 			self.MakePeerConn(u.Id, true)
	// 		}
	// 	}
	// }

	// return nil
}