package core

import (
	"bytes"
	"errors"
	"io"
	"strconv"
	"time"
)

var (
	RESP_NIL     []byte = []byte("$-1\r\n")
	RESP_OK      []byte = []byte("+OK\r\n")
	RESP_ZERO    []byte = []byte(":0\r\n")
	RESP_ONE     []byte = []byte(":1\r\n")
	RESP_MINUS_1 []byte = []byte(":-1\r\n")
	RESP_MINUS_2 []byte = []byte(":-2\r\n")
)

func evalPing(cmdArgs []string) []byte {
	var b []byte
	if len(cmdArgs) >= 2 {
		return Encode(errors.New("ERR wrong number of arguments for 'ping' command"), false)
	}
	if len(cmdArgs) == 0 {
		b = Encode("PONG", true)
	} else {
		b = Encode(cmdArgs[0], false)
	}
	return b
}

func evalGET(cmdArgs []string) []byte {
	if len(cmdArgs) != 1 {
		return Encode(errors.New("(error) ERR wrong number of arguments for 'get' command"), false)
	}
	key := cmdArgs[0]

	// get the value from the hash table
	obj := Get(key)

	// if key does not exist,return RESP encoded nil
	if obj == nil {
		return RESP_NIL
	}

	// if key already expired then return nil
	if obj.ExpiresAt != -1 && obj.ExpiresAt <= time.Now().UnixMilli() {
		return RESP_NIL
	}

	// return the RESP encoded value
	return Encode(obj.Value, false)
}

func evalSET(cmdArgs []string) []byte {
	if len(cmdArgs) <= 1 {
		return Encode(errors.New("(error) ERR wrong number of arguments for 'set' command"), false)
	}
	var exDurationMs int64 = -1
	key, val := cmdArgs[0], cmdArgs[1]
	for i := 2; i < len(cmdArgs); i++ {
		switch cmdArgs[i] {
		case "EX", "ex":
			i++
			if i == len(cmdArgs) {
				return Encode(errors.New("(error) ERR syntax error"), false)
			}
			exDurationSec, err := strconv.ParseInt(cmdArgs[3], 10, 64)
			if err != nil {
				return Encode(errors.New("(error) ERR value is not an integer or out of range "), false)
			}
			exDurationMs = exDurationSec * 1000
		default:
			return Encode(errors.New("(error) ERR syntax error"), false)
		}
	}

	// putting the key and value in a hash table
	Put(key, NewObj(val, exDurationMs))
	return RESP_OK
}

func evalTTL(cmdArgs []string) []byte {
	if len(cmdArgs) != 1 {
		return Encode(errors.New("(error) ERR wrong number of arguments for 'ttl' command"), false)
	}
	key := cmdArgs[0]

	obj := Get(key)

	// if key does not exist in the hash map, return RESP encoded -2 denoting the key does not exist
	if obj == nil {
		return RESP_MINUS_2
	}

	// if object exists, but no expiration is set then send -1
	if obj.ExpiresAt == -1 {
		return RESP_MINUS_1
	}

	// compute the time remaining for the key to expire and
	// return the RESP encoded form of it
	durTTL := obj.ExpiresAt - time.Now().UnixMilli()

	// if key expired i.e. durTTL <=0 key does not exist return -2
	if durTTL < 0 {
		return RESP_MINUS_2
	}

	return Encode(int64(durTTL/1000), false)
}

func evalDEL(cmdArgs []string) []byte {
	deleted := 0
	for _, key := range cmdArgs {
		ok := Del(key)
		if ok {
			deleted++
		}
	}
	return Encode(deleted, false)
}

func evalEXPIRE(cmdArgs []string) []byte {
	if len(cmdArgs) != 2 {
		return Encode(errors.New("(error) ERR wrong number of arguments for 'expire' command"), false)
	}
	key := cmdArgs[0]
	exDurationSec, err := strconv.ParseInt(cmdArgs[1], 10, 64)
	if err != nil {
		return Encode(errors.New("(error) ERR value is not an integer or out of range "), false)
	}
	val := Get(key)
	if val == nil {
		// key does not exist in the hashmap
		return Encode(0, false)
	}
	val.ExpiresAt = time.Now().UnixMilli() + exDurationSec*1000
	return Encode(1, false)
}

func EvalAndRespond(cmds RedisCmds, cli io.ReadWriter) {
	// Buffer responses
	var response []byte
	buf := bytes.NewBuffer(response)

	for _, cmd := range cmds {
		// evaluate all cmds one by one
		switch cmd.Cmd {
		case "PING":
			buf.Write(evalPing(cmd.Args))
		case "GET":
			buf.Write(evalGET(cmd.Args))
		case "SET":
			buf.Write(evalSET(cmd.Args))
		case "TTL":
			buf.Write(evalTTL(cmd.Args))
		case "DEL":
			buf.Write(evalDEL(cmd.Args))
		case "EXPIRE":
			buf.Write(evalEXPIRE(cmd.Args))
		default:
			buf.Write(evalPing(cmd.Args))
		}
	}
	cli.Write(buf.Bytes())
}
