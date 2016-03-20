package main

import (
	"time"
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