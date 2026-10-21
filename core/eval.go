package core

import (
	"errors"
	"net"
)

func evalPing(cmdArgs []string, cli net.Conn) error {
	var b []byte
	if len(cmdArgs) >= 2 {
		return errors.New("ERR wrong number of arguments for 'ping' command")
	}
	if len(cmdArgs) == 0 {
		b = Encode("PONG", true)
	} else {
		b = Encode(cmdArgs[0], false)
	}
	_, err := cli.Write(b)
	return err
}

func EvalAndRespond(cmd *RedisCmd, cli net.Conn) error {
	switch cmd.Cmd {
	case "PING":
		return evalPing(cmd.Args, cli)
	default:
		return evalPing(cmd.Args, cli)
	}
}
