package core

import "time"

var store map[string]*Obj

type Obj struct {
	Value     any
	ExpiresAt int64
}

func init() {
	store = make(map[string]*Obj)
}

func NewObj(val any, t int64) *Obj {
	var expiresAt int64 = -1
	if t > 0 {
		expiresAt = time.Now().UnixMilli() + t
	}
	return &Obj{
		Value:     val,
		ExpiresAt: expiresAt,
	}
}

func Put(key string, obj *Obj) {
	store[key] = obj
}

func Get(key string) *Obj {
	return store[key]
}
