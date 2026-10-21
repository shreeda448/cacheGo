package core

import (
	"errors"
	"fmt"
)

// reads the length typically the first integer of the string
// until hit by an non-digit byte and returns
// the integer and the delta = length + 2(CRLF)
func readLength(data []byte) (int, int) {
	length := 0
	for pos := range data {
		b := data[pos]
		if !(b >= '0' && b <= '9') {
			return length, pos + 2
		}
		length = length*10 + int(b-'0')
	}
	return 0, 0
}

// reads a RESP encoded simple string from data and
// returns the string,the delta and the error
func readSimpleString(data []byte) (string, int, error) {
	// first character +
	pos := 1
	for ; data[pos] != '\r'; pos++ {
	}
	return string(data[1:pos]), pos + 2, nil
}

// reads a RESP encoded error from data and returns
// the error string,the delta and the error
func readError(data []byte) (string, int, error) {
	return readSimpleString(data)
}

// reads a RESP encoded integer from data and returns
// the integer,the delta and the error
func readInt64(data []byte) (int64, int, error) {
	// first character :
	pos := 1
	var val int64 = 0
	for ; data[pos] != '\r'; pos++ {
		val = val*10 + int64(data[pos]-'0')
	}
	return val, pos + 2, nil
}

// reads a RESP encoded bulkstring from data and returns
// the string,the delta and the error
func readBulkString(data []byte) (string, int, error) {
	// first character $
	pos := 1
	// reading the length and forwarding the pos by the length of the integer + first special character
	len, delta := readLength(data[pos:])
	pos += delta
	// reading 'len' bytes as string
	return string(data[pos:(pos + len)]), pos + len + 2, nil
}

// reads RESP encoded array from data and returns
// the array,the delta and the error
func readArray(data []byte) (any, int, error) {
	// first character *
	pos := 1
	// reading the length
	cnt, delta := readLength(data[pos:])
	pos += delta
	elems := make([]any, cnt)
	for i := range elems {
		elem, delta, err := DecodeOne(data[pos:])
		if err != nil {
			return nil, 0, err
		}
		elems[i] = elem
		pos += delta
	}
	return elems, pos, nil
}

func DecodeArrayString(data []byte) ([]string, error) {
	val, err := Decode(data)
	if err != nil {
		return nil, err
	}
	temp := val.([]any)
	tokens := make([]string, len(temp))
	for i := range tokens {
		tokens[i] = temp[i].(string)
	}
	return tokens, nil
}

func DecodeOne(data []byte) (any, int, error) {
	if len(data) == 0 {
		return nil, 0, errors.New("no data")
	}
	switch data[0] {
	case '+':
		return readSimpleString(data)
	case '-':
		return readError(data)
	case ':':
		return readInt64(data)
	case '$':
		return readBulkString(data)
	case '*':
		return readArray(data)
	}
	return nil, 0, nil
}

func Decode(data []byte) (any, error) {
	if len(data) == 0 {
		return nil, errors.New("no data")
	}
	value, _, err := DecodeOne(data)
	return value, err
}

func Encode(val any, isSimple bool) []byte {
	switch v := val.(type) {
	case string:
		if isSimple {
			return fmt.Appendf(nil, "+%s\r\n", v)
		}
		return fmt.Appendf(nil, "$%d\r\n%s\r\n", len(v), v)
	}
	return []byte{}
}
