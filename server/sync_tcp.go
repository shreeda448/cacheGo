package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/shreeda448/cacheGo.git/config"
	"github.com/shreeda448/cacheGo.git/core"
)

func respondError(err error, cli net.Conn) {
	cli.Write([]byte(fmt.Sprintf("-%s\r\n", err)))
}

func respond(cmd *core.RedisCmd, cli net.Conn) {
	err := core.EvalAndRespond(cmd, cli)
	if err != nil {
		respondError(err, cli)
	}
}

func readCommand(cli net.Conn) (*core.RedisCmd, error) {
	// TODO: Max read in one shot is 512 bytes
	// To allow input > 512 bytes , then repeated read until
	// we get EOF or designated delimiter

	buf := make([]byte, 512)
	n, err := cli.Read(buf[:])
	if err != nil {
		return nil, err
	}
	tokens, err := core.DecodeArrayString(buf[:n])
	if err != nil {
		return nil, err
	}
	return &core.RedisCmd{
		Cmd:  strings.ToUpper(tokens[0]),
		Args: tokens[1:],
	}, nil
}

func RunSyncTCPServer() {
	log.Println("starting a synchronous TCP server on", config.Host, config.Port)
	numClients := 0
	// listening to the configured host:port
	listener, err := net.Listen("tcp", config.Host+":"+strconv.Itoa(config.Port))
	if err != nil {
		panic(err)
	}
	for {
		// blocking call: waiting for the client to connect
		cli, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		// increment the number of concurrent clients
		numClients++
		log.Println("client connected with address:", cli.RemoteAddr(), "concurrent clients", numClients)
		for {
			// over the socket, continously read the command and print it out
			cmd, err := readCommand(cli)
			if err != nil {
				cli.Close()
				numClients--
				log.Println("client disconnected", cli.RemoteAddr(), "concurrent clients", numClients)
				if err == io.EOF {
					break
				}
				log.Println("err", err)
			}
			respond(cmd, cli)
		}
	}
}
