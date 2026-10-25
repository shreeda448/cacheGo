package core

import (
	"errors"
	"io"
	"strconv"
	"time"
)

var RESP_NIL []byte = []byte("$-1\r\n")

func evalPing(cmdArgs []string, cli io.ReadWriter) error {
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

func evalGET(cmdArgs []string, cli io.ReadWriter) error {
	if len(cmdArgs) != 1 {
		return errors.New("(error) ERR wrong number of arguments for 'get' command")
	}
	key := cmdArgs[0]

	// get the value from the hash table
	obj := Get(key)

	// if key does not exist,return RESP encoded nil
	if obj == nil {
		cli.Write(RESP_NIL)
		return nil
	}

	// if key already expired then return nil
	if obj.ExpiresAt != -1 && obj.ExpiresAt <= time.Now().UnixMilli() {
		cli.Write(RESP_NIL)
		return nil
	}

	// return the RESP encoded value
	cli.Write(Encode(obj.Value, false))
	return nil
}

func evalSET(cmdArgs []string, cli io.ReadWriter) error {
	if len(cmdArgs) <= 1 {
		return errors.New("(error) ERR wrong number of arguments for 'set' command")
	}
	var exDurationMs int64 = -1
	key, val := cmdArgs[0], cmdArgs[1]
	for i := 2; i < len(cmdArgs); i++ {
		switch cmdArgs[i] {
		case "EX", "ex":
			i++
			if i == len(cmdArgs) {
				return errors.New("(error) ERR syntax error")
			}
			exDurationSec, err := strconv.ParseInt(cmdArgs[3], 10, 64)
			if err != nil {
				return errors.New("(error) ERR value is not an integer or out of range ")
			}
			exDurationMs = exDurationSec * 1000
		default:
			return errors.New("(error) ERR syntax error")
		}
	}

	// putting the key and value in a hash table
	Put(key, NewObj(val, exDurationMs))
	cli.Write([]byte("+OK\r\n"))
	return nil
}

func evalTTL(cmdArgs []string, cli io.ReadWriter) error {
	if len(cmdArgs) != 1 {
		return errors.New("(error) ERR wrong number of arguments for 'ttl' command")
	}
	key := cmdArgs[0]

	obj := Get(key)

	// if key does not exist in the hash map, return RESP encoded -2 denoting the key does not exist
	if obj == nil {
		cli.Write([]byte(":-2\r\n"))
		return nil
	}

	// if object exists, but no expiration is set then send -1
	if obj.ExpiresAt == -1 {
		cli.Write([]byte(":-1\r\n"))
		return nil
	}

	// compute the time remaining for the key to expire and
	// return the RESP encoded form of it
	durTTL := obj.ExpiresAt - time.Now().UnixMilli()

	// if key expired i.e. durTTL <=0 key does not exist return -2
	if durTTL < 0 {
		cli.Write([]byte(":-2\r\n"))
		return nil
	}

	cli.Write(Encode(int64(durTTL/1000), false))
	return nil
}

func EvalAndRespond(cmd *RedisCmd, cli io.ReadWriter) error {
	switch cmd.Cmd {
	case "PING":
		return evalPing(cmd.Args, cli)
	case "GET":
		return evalGET(cmd.Args, cli)
	case "SET":
		return evalSET(cmd.Args, cli)
	case "TTL":
		return evalTTL(cmd.Args, cli)
	default:
		return evalPing(cmd.Args, cli)
	}
}
