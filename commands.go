package main

import (
	"bytes"
	"reflect"
	"strconv"
	"time"
	"unicode"
)

type result struct {
	value []byte
	err   error
}

type commands struct {
	cache *cache
}

type ProtocolError struct {
}

type WrongCommandError struct {
}

type EmptyCommandError struct {
}

type tokens [][]byte

func (e ProtocolError) Error() string {
	return "protoErr"
}

func (e WrongCommandError) Error() string {
	return "wrongCommand"
}

func (e EmptyCommandError) Error() string {
	return "emptyCommand"
}

func (c *commands) run(line []byte) *result {
	r := &result{}
	tokens := c.parse(line)

	if len(tokens) == 0 {
		r.err = EmptyCommandError{}
		return r
	}

	command := bytes.ToLower(tokens[0])
	first := bytes.ToUpper(command[0:1])
	rest := command[1:]
	command = bytes.Join([][]byte{first, rest}, nil)

	method := reflect.ValueOf(c).MethodByName("Command" + string(command))

	if method.IsValid() {
		return method.Call([]reflect.Value{reflect.ValueOf(tokens[1:])})[0].Interface().(*result)
	}

	r.err = WrongCommandError{}

	return r
}

func (c *commands) parse(line []byte) [][]byte {
	tokens := [][]byte{}

	var isQuote bool
	buffer := &bytes.Buffer{}

	for _, bt := range line {
		r := rune(bt)

		if unicode.IsSpace(r) && !isQuote {
			if buffer.Len() > 0 {
				tokens = append(tokens, buffer.Bytes())
			}

			buffer = &bytes.Buffer{}
		} else {
			if r == rune([]byte(`"`)[0]) {
				if isQuote {
					isQuote = false
					tokens = append(tokens, buffer.Bytes())
					buffer = &bytes.Buffer{}
				} else {
					isQuote = true
				}

				continue
			}

			buffer.WriteRune(r)
		}
	}

	return tokens
}

// Command "DEL key"
func (c *commands) CommandDel(t tokens) *result {
	r := &result{}

	if len(t) < 1 {
		r.err = ProtocolError{}
		return r
	}

	key := string(t[0])

	if c.cache.delete(key) {
		r.value = []byte("true")
	} else {
		r.value = []byte("false")
	}

	return r
}

// Command "TOUCH key"
func (c *commands) CommandTouch(t tokens) *result {
	r := &result{}

	if len(t) < 1 {
		r.err = ProtocolError{}
		return r
	}

	key := t[0]
	entry := c.cache.get(string(key))

	if entry == nil {
		r.value = []byte("false")
		return r
	}

	entry.lastTouchedAt = time.Now()
	r.value = []byte("true")

	return r
}

// Command "SETX key value"
func (c *commands) CommandSetx(t tokens) *result {
	r := &result{}

	if len(t) < 2 {
		r.err = ProtocolError{}
		return r
	}

	key := t[0]
	value, _ := strconv.Atoi(string(t[1]))
	entry := c.cache.get(string(key))

	if entry == nil || value < 0 {
		r.value = []byte("false")
		return r
	}

	entry.ttl = value
	r.value = []byte("true")

	return r
}

// Command "TTL key"
func (c *commands) CommandTtl(t tokens) *result {
	r := &result{}

	if len(t) < 1 {
		r.err = ProtocolError{}
		return r
	}

	key := string(t[0])

	if !c.cache.exists(key) {
		r.value = []byte("nil")
		return r
	}

	r.value = []byte(strconv.Itoa(c.cache.getTTL(key)))

	return r
}

// Command "DECR key value"
func (c *commands) CommandDecr(t tokens) *result {
	r := &result{}

	if len(t) < 1 {
		r.err = ProtocolError{}
		return r
	}

	key := string(t[0])
	delta := 1

	if len(t) >= 2 {
		delta, _ = strconv.Atoi(string(t[1]))
	}

	if delta <= 0 {
		delta = 1
	}

	c.cache.incr(key, -delta)

	return c.CommandGet(t)
}

// Command "INCR key value"
func (c *commands) CommandIncr(t tokens) *result {
	r := &result{}

	if len(t) < 1 {
		r.err = ProtocolError{}
		return r
	}

	key := string(t[0])
	delta := 1

	if len(t) >= 2 {
		delta, _ = strconv.Atoi(string(t[1]))
	}

	if delta <= 0 {
		delta = 1
	}

	c.cache.incr(key, delta)

	return c.CommandGet(t)
}

// Command "ADD key value ttl"
func (c *commands) CommandAdd(t tokens) *result {
	r := &result{}

	if len(t) < 1 {
		r.err = ProtocolError{}
		return r
	}

	key := string(t[0])

	if c.cache.exists(key) {
		r.value = []byte("nil")
		return r
	}

	return c.CommandSet(t)
}

// Command "EXISTS key"
func (c *commands) CommandExists(t tokens) *result {
	r := &result{}

	if len(t) < 1 {
		r.err = ProtocolError{}
		return r
	}

	key := string(t[0])

	if c.cache.exists(key) {
		r.value = []byte("true")
	} else {
		r.value = []byte("false")
	}

	return r
}

// Command "GET key"
func (c *commands) CommandGet(t tokens) *result {
	r := &result{}

	if len(t) < 1 {
		r.err = ProtocolError{}
		return r
	}

	key := t[0]
	entry := c.cache.get(string(key))

	if entry != nil {
		r.value = entry.value
	} else {
		r.value = []byte("nil")
	}

	return r
}

// Command "SET key value TTL"
func (c *commands) CommandSet(t tokens) *result {
	r := &result{}

	if len(t) < 2 {
		r.err = ProtocolError{}
		return r
	}

	key := t[0]
	ttl := 0

	if len(t) >= 3 {
		ttl, _ = strconv.Atoi(string(t[2]))
	}

	c.cache.set(string(key), t[1], ttl)
	r.value = t[1]

	return r
}
