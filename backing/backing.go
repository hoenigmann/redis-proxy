package backing

import (
	"github.com/mediocregopher/radix.v2/pool"
	"github.com/mediocregopher/radix.v2/redis"
)

type Backing struct {
	host string
	port string
	pool *pool.Pool
}

func New(host string, port string) *Backing {
	b := new(Backing)
	b.host = host
	b.port = port
	connPool, err := pool.New("tcp", b.host+":"+b.port, 10)
	if err != nil {
		panic(err)
	}
	b.pool = connPool
	return b
}

func (b *Backing) Get(key string) (string, bool) {
	conn, err := b.pool.Get()
	defer b.pool.Put(conn)
	if err != nil {
		panic(err)
	}
	resp := conn.Cmd("GET", key)
	if resp.Err != nil {
		panic(err)
	} else {
		if resp.IsType(redis.Int | redis.Array | redis.Nil | redis.SimpleStr | redis.BulkStr) {
			bytes, err := resp.Bytes()
			if err != nil {
				return "", false
			} else {
				return string(bytes), true
			}
		}
	}
	return "", false
}

func (b *Backing) Close() {
	b.pool.Empty()
}
