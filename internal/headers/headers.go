package headers

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

type Headers map[string]string

const crlf = "\r\n"

func NewHeaders() Headers {
	return map[string]string{}
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte(crlf))

	if idx == -1 {
		return 0, false, nil
	}

	if idx == 0 {
		return 2, true, nil
	}

	parts := bytes.SplitN(data[:idx], []byte(":"), 2)
	key := string(parts[0])

	if key != strings.TrimRight(key, " ") {
		return 0, false, fmt.Errorf("invalid header name: %s", key)
	}

	value := bytes.TrimSpace(parts[1])
	key = strings.TrimSpace(key)

	if !isValidKey(key) {
		return 0, false, fmt.Errorf("invalid parameter")
	}
	h.Set(key, string(value))

	return idx + 2, false, nil
}

func (h Headers) Set(key, value string) {
	key = strings.ToLower(key)

	if _, exist := h[key]; exist {
		h[key] += ", " + value
	} else {
		h[key] = value
	}
}

func (h Headers) Get(key string) (string, bool) {
	key = strings.ToLower(key)
	v, ok := h[key]
	return v, ok
}

func isValidKey(key string) bool {
	pattern := `^[A-Za-z0-9!#$%&'*+\-.\^_` + "`" + `|~]+$`
	reg := regexp.MustCompile(pattern)

	return reg.MatchString(key)
}
