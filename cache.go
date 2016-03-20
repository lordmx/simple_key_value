package main

import (
	"time"
	"log"
	"strconv"
)

type entry struct {
	id int
	key string
	value []byte
	lastTouchedAt time.Time
	ttl int
}

type cache struct {
	entriesCount int
	entries map[string]*entry
	freeList map[int][]*entry
	freeListTs []int
	ticker *time.Ticker
	done chan bool 
}

func NewCache() *cache {
	c := &cache{
		entries: make(map[string]*entry),
		freeList: make(map[int][]*entry),
		done: make(chan bool),
	}

	c.ticker = time.NewTicker(1 * time.Second)
	go c.expires()

	return c
}

func (cache *cache) expires() bool {
	for {
		select {
		case <- cache.ticker.C:
			ts := int(time.Now().Unix())

			if entries, ok := cache.freeList[ts]; ok {
				for _, entry := range entries {
					cache.delete(entry.key)
				}

				log.Printf("[%v] Deleted %d items", time.Now(), len(cache.freeList[ts]))
				
				delete(cache.freeList, ts)
				
			}

			for i, value := range cache.freeListTs {
				if value == ts {
					break
				}

				cache.freeListTs = append(cache.freeListTs[:i], cache.freeListTs[i+1:]...)
			}

			if len(cache.freeListTs) > 1 {
				cache.freeListTs = cache.freeListTs[1:]
			} else {
				cache.freeListTs = []int{}
			}
		case d := <- cache.done:
			if d {
				cache.ticker.Stop()
				return true
			}
		}
	}

	return false
}

func (cache *cache) exists(key string) bool {
	if _, ok := cache.entries[key]; ok {
		return true
	}

	return false
}

func (cache *cache) delete(key string) bool {
	if cache.exists(key) {
		delete(cache.entries, key)
		cache.entriesCount--

		return true
	}

	return false
}

func (cache *cache) setTTL(key string, TTL int) bool {
	if e, ok := cache.entries[key]; ok {
		e.ttl = TTL
		e.lastTouchedAt = time.Now()

		if TTL > 0 {
			expires := int(e.lastTouchedAt.Unix()) + TTL

			cache.freeList[expires] = append(cache.freeList[expires], e)
			cache.freeListTs = append(cache.freeListTs, expires)
		}

		return true
	}

	return false
}

func (cache *cache) getTTL(key string) int {
	if e, ok := cache.entries[key]; ok {
		if e.ttl == 0 {
			return -1
		}

		return int((e.lastTouchedAt.Unix() + int64(e.ttl)) - time.Now().Unix())
	}

	return 0
}

func (cache *cache) incr(key string, delta ...int) *entry {
	if !cache.exists(key) {
		cache.set(key, []byte("0"))		
	}

	if len(delta) == 0 {
		delta[0] = 1
	}

	value, _ := strconv.Atoi(string(cache.get(key).value))
	cache.set(key, []byte(strconv.Itoa(value + delta[0])))

	return cache.get(key)
}

func (cache *cache) get(key string) *entry {
	if e, ok := cache.entries[key]; ok {
		if e.ttl > 0 && cache.getTTL(key) <= 0 {
			cache.delete(key)
			return nil
		}

		return e
	}

	return nil
}

func (cache *cache) set(key string, value []byte, ttl ...int) {
	var e *entry

	if cache.exists(key) {
		e = cache.get(key)
	} else {
		cache.entriesCount++

		e = &entry{
			key: key,
		}

		cache.entries[key] = e
	}

	e.value = value

	if len(ttl) == 0 {
		ttl[0] = 0
	}

	if ttl[0] < 0 {
		ttl[0] = 0
	}

	cache.setTTL(key, ttl[0])
}