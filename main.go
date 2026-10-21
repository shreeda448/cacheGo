package main

import (
	"flag"
	"log"

	"github.com/shreeda448/cacheGo.git/config"
	"github.com/shreeda448/cacheGo.git/server"
)

func setupFlags() {
	flag.StringVar(&config.Host, "host", "0.0.0.0", "host for the dice server")
	flag.IntVar(&config.Port, "port", 7379, "port for the dice server")
	flag.Parse()
}

func main() {
	setupFlags()
	log.Println("Let's Go, cacheGo")
	server.RunSyncTCPServer()
}
