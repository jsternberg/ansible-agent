package main

import (
	"crypto/tls"
	"flag"
	"io/ioutil"
	"log"
	"net"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/jsternberg/ansible-agent/ansible"
)

var (
	flConfig = flag.String("c", "", "Server configuration file")
)

func realMain() int {
	flag.Parse()

	config := DefaultConfig()
	if *flConfig != "" {
		in, err := ioutil.ReadFile(*flConfig)
		if err != nil {
			log.Println(err)
			return 1
		}

		if err := toml.Unmarshal(in, config); err != nil {
			log.Println(err)
			return 1
		}
	}

	l, err := net.Listen("tcp", ":8700")
	if err != nil {
		log.Println(err)
		return 1
	}

	if config.SSL.Enabled {
		cert, err := tls.LoadX509KeyPair(config.SSL.Certificate, config.SSL.PrivateKey)
		if err != nil {
			log.Println(err)
			return 1
		}

		tlsConfig := tls.Config{
			Certificates: []tls.Certificate{cert},
		}
		l = tls.NewListener(l, &tlsConfig)
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
