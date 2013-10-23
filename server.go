package main

import (
	"encoding/gob"
	"log"
	"net"
	"./utils"
)

var users map[string]*User

func runTracker(host string) error {
	users = make(map[string]*User)

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

		u := &User{
			conn:    conn,
			encoder: gob.NewEncoder(conn),
			decoder: gob.NewDecoder(conn),
		}
		
		go runListener(u, users)


	}

	return nil
}

func sendState(u *User, users map[string]*User) {
	// You just connected, let's tell you about the other users and you can connect to them
	var userList []User
	for _, iter_user := range users {
		if iter_user.Id != u.Id { // ignore ourself (the client doesn't know their own ID because this is a trivial example app)
			userList = append(userList, *iter_user)
		}
	}
	env := &Envelope{
		UserList: userList,
	}
	err := u.encoder.Encode(env)
	if err != nil {
		// TODO: retry sending user list
	}
}

func runListener(u *User, users map[string]*User) {
	for {
		env := new(Envelope)
		err := u.decoder.Decode(&env)
		if err != nil {
			log.Println("forgetting user: ", u.Id)
			// remove this user/conn from the map?
			return
		}

		if env.Auth != nil {
			// Create new user
			if env.Auth.UUID == "" {
				env.Auth.UUID, _ = utils.GenUUID()
				u.Id = env.Auth.UUID
				u.encoder.Encode(env)
				log.Println("Sent UUID:", u.Id)
				users[u.Id] = u

			// Authenticate
			} else {
				if user, ok := users[env.Auth.UUID]; ok {
					log.Println("Logging in UUID:", env.Auth.UUID)
					user.conn = u.conn
					user.encoder = u.encoder
					user.decoder = u.decoder
				} else {
					log.Println("Unknown user tried login")
					return
				}
			}

			sendState(u, users)
		}

		if env.PcSignal != nil {
			env.PcSignal.From = u.Id

			log.Println("pcsignal", env.PcSignal.From, "->", env.PcSignal.To)

			toUser := users[env.PcSignal.To]
			if toUser != nil {
				toUser.encoder.Encode(env)
			}
		}
	}
}