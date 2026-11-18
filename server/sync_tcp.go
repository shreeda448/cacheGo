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

func respondError(err error, cli io.ReadWriter) {
	_, _ = fmt.Fprintf(cli, "-%s\r\n", err)
}

func respond(cmd core.RedisCmds, cli io.ReadWriter) {
	core.EvalAndRespond(cmd, cli)
}

func toArrayString(ai []any) ([]string, error) {
	as := make([]string, len(ai))
	for i := range ai {
		as[i] = ai[i].(string)
	}
	return as, nil
}

func readCommands(cli io.ReadWriter) (core.RedisCmds, error) {
	// TODO: Max read in one shot is 512 bytes
	// To allow input > 512 bytes , then repeated read until
	// we get EOF or designated delimiter

	buf := make([]byte, 512)
	n, err := cli.Read(buf[:])
	if err != nil {
		return nil, err
	}
	values, err := core.Decode(buf[:n])
	if err != nil {
		return nil, err
	}
	cmds := make([]*core.RedisCmd, 0)
	for _, value := range values {
		tokens, err := toArrayString(value.([]any))
		if err != nil {
			return nil, err
		}
		cmds = append(cmds, &core.RedisCmd{
			Cmd:  strings.ToUpper(tokens[0]),
			Args: tokens[1:],
		})
	}
	return cmds, nil
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
			cmd, err := readCommands(cli)
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
