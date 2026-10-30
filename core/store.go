package core

import (
	"time"

	"github.com/shreeda448/cacheGo.git/config"
)

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
	if len(store) >= config.KeysLimit {
		evict()
	}
	store[key] = obj
}

func Get(key string) *Obj {
	val := store[key]
	if val != nil && val.ExpiresAt != -1 && val.ExpiresAt <= time.Now().UnixMilli() {
		delete(store, key)
		return nil
	}
	return val
}

func Del(key string) bool {
	_, ok := store[key]
	if ok {
		delete(store, key)
		return true
	}
	return false
}
