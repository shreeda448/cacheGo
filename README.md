# CacheGo

**CacheGo** is my implementation of redis-like cache db in golang . I have built this purely to learn more about redis and async programming

## Build

```
go build
```

### Currently implemented features

- RESP like datatype parsing
- PING command
- Async I/O using `epoll`
- GET command
- SET command with expiration time flag
- TTL command
- DEL command
- Simple Key eviction strategy
- Expire command
- Auto deletion after expiration
- Lazy deletion

### Future Features

- Pub Sub
- Atomic operations/Transactions
- LRU and LFU key eviction strategies
