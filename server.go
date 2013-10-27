package main

import (
	"encoding/gob"
	"log"
	"net"
	"./utils"
)

var users map[string]*Tracker

func runTracker(host string) error {
	users = make(map[string]*Tracker)

	addr, err := net.ResolveTCPAddr("tcp", host)
	if err != nil {
		return err
	}
	log.Println("Tracker ready on:", addr)

	conn, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}

	for {
		conn, err := conn.Accept()
		if err != nil {
			log.Println("Error accepting:", err)
			continue
		}

		t := &Tracker{
			conn:    conn,
			encoder: gob.NewEncoder(conn),
			decoder: gob.NewDecoder(conn),
		}
		
		go runListener(t, users)

	}

	return nil
}

func sendState(t *Tracker, users map[string]*Tracker) {
	// You just connected, let's tell you about the other users and you can connect to them
	var userList []Tracker
	for _, user := range users {
		if user.UUID != t.UUID { // ignore ourself (the client doesn't know their own ID because this is a trivial example app)
			userList = append(userList, *user)
			log.Println("Added remote user:", user.UUID, t.UUID)
		}
	}
	env := &Envelope{
		UserList: userList,
	}
	err := t.encoder.Encode(env)
	if err != nil {
		// TODO: retry sending user list
	}
}

func runListener(t *Tracker, users map[string]*Tracker) {
	for {
		env := new(Envelope)
		err := t.decoder.Decode(&env)
		if err != nil {
			log.Println("forgetting user: ", t.UUID)
			newusers := make(map[string]*Tracker)
			for _, user := range users {
				if user.UUID != t.UUID {
					newusers[user.UUID] = user
				}
			}
			users = newusers
			return
		}

		if env.Auth != nil {
			// Create new user
			if env.Auth.UUID == "" {
				env.Auth.UUID, _ = utils.GenUUID()
				t.UUID = env.Auth.UUID
				t.encoder.Encode(env)
				users[t.UUID] = t
				log.Println("Sent UUID:", t.UUID)

			// Authenticate
			} else {
				if user, ok := users[env.Auth.UUID]; ok {
					log.Println("Logging in UUID:", env.Auth.UUID)
					t.UUID = env.Auth.UUID
					user.UUID = env.Auth.UUID
					user.conn = t.conn
					user.encoder = t.encoder
					user.decoder = t.decoder
				} else {
					log.Println("Unknown user tried login")
					return
				}
			}

			sendState(t, users)
		}

		if env.PcSignal != nil {
			env.PcSignal.From = t.UUID

			log.Println("pcsignal", env.PcSignal.From, "->", env.PcSignal.To)

			toUser := users[env.PcSignal.To]
			if toUser != nil {
				toUser.encoder.Encode(env)
			}
		}
	}
}