package headers

import (
	"bytes"
	"fmt"
	"httpfromtcp/internal/constants"
	"strings"
	"unicode"
)

const colon = ':'
const ws = ' '
const allowedSpecialChracters = "!#$%&'*+-.^_`|~"

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	endIdx := bytes.Index(data, []byte(constants.CRLF))
	if endIdx == -1 {
		return 0, false, nil
	}
	if endIdx == 0 {
		return 2, true, nil
	}

	colonIdx := bytes.IndexRune(data, colon)
	if colonIdx == -1 {
		return 0, false, fmt.Errorf("malformed header: missing colon")
	}
	if colonIdx > 0 && data[colonIdx-1] == ws {
		return 0, false, fmt.Errorf("failed to parse header: extra whitespace before colon")
	}

	key := strings.ToLower(strings.TrimSpace(string(data[:colonIdx])))
	value := strings.TrimSpace(string(data[colonIdx+1 : endIdx]))

	if !h.isValidFieldName(key) {
		return 0, false, fmt.Errorf("failed to parse header: invalid character")
	}

	h.Set(key, value)

	return endIdx + 2, false, nil
}

func (h Headers) Set(key, value string) {
	key = strings.ToLower(key)
	v, ok := h[key]
	if ok {
		value = strings.Join([]string{v, value}, ", ")
	}
	h[key] = value
}

func (h Headers) Get(key string) (value string, ok bool) {
	value, ok = h[strings.ToLower(key)]

	return value, ok
}
func (h Headers) Delete(key string) {
	delete(h, strings.ToLower(key))
}

func (h Headers) Override(key, value string) {
	key = strings.ToLower(key)
	h[key] = value
}

func (h Headers) isValidFieldName(fieldName string) bool {
	for _, ch := range fieldName {
		isLetter := unicode.IsLetter(ch)
		isDigit := unicode.IsDigit(ch)
		isAllowedSpecialCharacter := strings.ContainsRune(allowedSpecialChracters, ch)

		if !isLetter && !isDigit && !isAllowedSpecialCharacter {
			return false
		}
	}

	return true
}
