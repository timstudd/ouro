package main

import (
	"flag"
)

var (
	tracker_flag = flag.String("tracker", "", "Port/interface to listen on")
	client_flag = flag.String("client", "", "Host/port to connect to")
	default_tracker = "hdcserver.centrapi.com:9000"
)

type Envelope struct {
	PcSignal	*PcSignal
	UserList	[]Tracker
	Auth		*Auth
}

type Auth struct {
	UUID	string	
}

type PcSignal struct {
	From    string
	To      string
	Payload []byte
}

func init() {
	flag.Parse()
}

func main() {
	// Use default tracker as client if one is not provided
	if *tracker_flag == "" && *client_flag == "" {
		client_flag = &default_tracker
	}

	if *tracker_flag != "" {
		err := runTracker(*tracker_flag)
		if err != nil {
			panic(err)
		}
	} else if *client_flag != "" {
		err := runClient(*client_flag)
		if err != nil {
			panic(err)
		}
	}
}
