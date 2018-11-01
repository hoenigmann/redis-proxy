package lru

import (
	"container/list"
	"sync"
	"time"
)

type Key interface{}

type entry struct {
	key        Key
	value      string
	expiration time.Time
}

type Cache struct {
	items        map[Key]*list.Element
	mu           sync.RWMutex
	lruqueue     *list.List
	MaxEntries   int
	globalExpiry time.Duration
}

func New(capacity int, globalExpiry time.Duration) *Cache {
	c := new(Cache)
	c.items = make(map[Key]*list.Element)
	c.lruqueue = list.New()
	c.MaxEntries = capacity
	c.globalExpiry = globalExpiry
	return c
}

func (c *Cache) Get(key Key) (string, bool) {
	c.mu.RLock()

	if val, ok := c.items[key]; ok {
		c.mu.RUnlock()
		if val.Value.(*entry).expiration.After(time.Now()) {
			c.mu.Lock()
			c.lruqueue.MoveToFront(val)
			//c.OntoFront<-val Should do this with buffered chang instead to avoid to much lock time on this thread
			c.mu.Unlock()
			return val.Value.(*entry).value, true
		} else { // Expired.
			c.mu.Lock()
			c.removeElement(val)
			c.mu.Unlock()
			return "", false
		}
	}
	c.mu.RUnlock()
	return "", false
}

/*
	Put adds the item to the q.
*/
func (c *Cache) Put(key string, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.items[key]; ok {
		// Its already there, guard.
		return
	}

	if c.MaxEntries == c.lruqueue.Len() {
		c.removeOldest()
	}

	expires := time.Now().Add(c.globalExpiry)
	element := c.lruqueue.PushFront(&entry{key, value, expires})
	c.items[key] = element
}

func (c *Cache) removeOldest() {
	element := c.lruqueue.Back()
	if element != nil {
		c.removeElement(element)
	}
}

func (c *Cache) removeElement(element *list.Element) {
	c.lruqueue.Remove(element)
	entr := element.Value.(*entry)
	delete(c.items, entr.key)
}
