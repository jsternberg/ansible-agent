package main

import (
	"log"
	"net"
	"os"

	"github.com/jsternberg/ansible-agent/ansible"
)

func realMain() int {
	l, err := net.Listen("tcp", ":8700")
	if err != nil {
		log.Println(err)
		return 1
	}

	server := ansible.NewServer()
	if err := server.Serve(l); err != nil {
		log.Println(err)
		return 1
	}
	return 0
}

func main() {
	os.Exit(realMain())
}
