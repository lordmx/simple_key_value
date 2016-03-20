package main

import (
	"reflect"
	"bytes"
	"strconv"
	"unicode"
)

type result struct {
	value []byte
	err error
}

type commands struct {
	cache *cache
}

type ProtocolError struct {

}

type WrongKeyError struct {

}

type WrongCommandError struct {

}

type EmptyCommandError struct {

}

func (e ProtocolError) Error() string {
	return "protoerr"
}

func (e WrongKeyError) Error() string {
	return "wrongkey"
}

func (e WrongCommandError) Error() string {
	return "wrongcommand"
}

func (e EmptyCommandError) Error() string {
	return "emptycommand"
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

func (c *commands) CommandGet(tokens[][]byte) *result {
	r := &result{}

	if len(tokens) < 1 {
		r.err = ProtocolError{}
		return r
	}

	key := tokens[0]

	if !unicode.IsLetter(rune(key[0])) {
		r.err = WrongKeyError{}
		return r
	}

	entry := c.cache.get(string(key))

	if entry != nil {
		r.value = entry.value
	} else {
		r.value = []byte("")
	}

	return r
}

func (c *commands) CommandSet(tokens[][]byte) *result {
	r := &result{}

	if len(tokens) < 2 {
		r.err = ProtocolError{}
		return r
	}

	key := tokens[0]

	if !unicode.IsLetter(rune(key[0])) {
		r.err = WrongKeyError{}
		return r
	}

	ttl := 0

	if len(tokens) == 3 {
		ttl, _ = strconv.Atoi(string(tokens[2]))
	}

	c.cache.set(string(key), tokens[1], ttl)
	r.value = tokens[1]

	return r
}