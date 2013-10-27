package main

import (
	"log"
	"time"
)

func runClient(host string) error {
	self := NewUser()

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
			go tracker.Listen(engine, self)
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
}