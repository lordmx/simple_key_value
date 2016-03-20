package main

import (
	"time"
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
}

func (cache *cache) exists(key string) bool {
	if _, ok := cache.entries[key]; ok {
		return true
	}

	return false
}

func (cache *cache) get(key string) *entry {
	if e, ok := cache.entries[key]; ok {
		now := time.Now()

		if e.ttl > 0 && (e.lastTouchedAt.Unix() + int64(e.ttl)) <= now.Unix() {
			delete(cache.entries, key)
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
			id: cache.entriesCount,
			key: key,
		}

		cache.entries[key] = e
	}

	e.value = value
	e.lastTouchedAt = time.Now()

	if len(ttl) > 0 && ttl[0] > 0 {
		e.ttl = ttl[0]
	}
}