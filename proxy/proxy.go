package proxy

import (
	"time"

	"github.com/hoenigmann/redis-proxy/backing"
	"github.com/hoenigmann/redis-proxy/lru"
)

type Proxy struct {
	store *backing.Backing
	cache *lru.Cache
}

func New(host string, port string, expiry time.Duration, capacity int) *Proxy {
	p := new(Proxy)
	p.store = backing.New(host, port)
	p.cache = lru.New(capacity, expiry)
	return p
}

func (p *Proxy) Get(key string) (string, bool) {
	if val, ok := p.cache.Get(key); ok {
		return val, true
	}

	if val, ok := p.store.Get(key); ok {
		p.cache.Put(key, val)
		return val, true
	}
	return "", false
}
