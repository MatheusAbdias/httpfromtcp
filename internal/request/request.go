package request

import (
	"bytes"
	"errors"
	"fmt"
	"httpfromtcp/internal/constants"
	"httpfromtcp/internal/headers"
	"io"
	"strconv"
	"strings"
	"unicode"
)

type RequestStatus int

const (
	Initialized RequestStatus = iota
	ParsingHeaders
	ParsingBody
	Done
)

const bufferSize = 8

func (r RequestStatus) String() string {
	return [...]string{"Initialized", "Done"}[r]
}

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	state       RequestStatus
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	var readToIndex int
	buf := make([]byte, bufferSize)
	req := &Request{
		state:   Initialized,
		Headers: headers.NewHeaders(),
		Body:    make([]byte, 0),
	}

	for req.state != Done {
		if readToIndex >= len(buf) {
			newBuff := make([]byte, len(buf)*2)
			copy(newBuff, buf)
			buf = newBuff
		}

		numBytesRead, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if req.state != Done {
					return nil, fmt.Errorf("incomplete request, in state: %d, read n bytes on EOF: %d", req.state, numBytesRead)
				}
				break
			}
			return nil, err
		}
		readToIndex += numBytesRead

		numBytesParsed, err := req.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[numBytesParsed:])
		readToIndex -= numBytesParsed

	}

	return req, nil
}

func (r *Request) parse(data []byte) (int, error) {
	var totalBytesParsed int
	for r.state != Done {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}
		totalBytesParsed += n
		if n == 0 {
			break
		}
	}

	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case Initialized:
		n, err := r.parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return 0, nil
		}
		r.state = ParsingHeaders
		return n, nil

	case ParsingHeaders:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if done {
			r.state = ParsingBody
		}
		return n, nil

	case ParsingBody:
		return r.parseBody(data)

	case Done:
		return 0, fmt.Errorf("error: trying to read data in a done state")

	default:
		return 0, fmt.Errorf("unknown state")
	}
}

func (r *Request) parseRequestLine(data []byte) (int, error) {
	idx := bytes.Index(data, []byte(constants.CRLF))
	if idx == -1 {
		return 0, nil
	}

	requestLineText := string(data[:idx])
	requestLine, err := r.requestLineFromString(requestLineText)
	if err != nil {
		return 0, err
	}
	r.RequestLine = *requestLine
	return idx + 2, nil
}

func (r *Request) requestLineFromString(str string) (*RequestLine, error) {
	parts := strings.Split(str, " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("poorly formatted request-line: %s", str)
	}

	method := parts[0]
	if !onlyCapitalLetters(method) {
		return nil, fmt.Errorf("invalid method: %s", method)
	}

	requestTarget := parts[1]
	versionParts := strings.Split(parts[2], "/")
	if len(versionParts) != 2 {
		return nil, fmt.Errorf("malformed start-line: %s", str)
	}

	httpPart := versionParts[0]
	if httpPart != "HTTP" {
		return nil, fmt.Errorf("unrecognized HTTP-version: %s", httpPart)
	}
	version := versionParts[1]
	if version != "1.1" {
		return nil, fmt.Errorf("unrecognized HTTP-version: %s", version)
	}
	return &RequestLine{
		Method:        method,
		RequestTarget: requestTarget,
		HttpVersion:   versionParts[1],
	}, nil
}

func onlyCapitalLetters(method string) bool {
	if len(method) == 0 {
		return false
	}
	for _, r := range method {
		if !unicode.IsLetter(r) || !unicode.IsUpper(r) {
			return false
		}
	}
	return true
}

func (r *Request) parseBody(data []byte) (int, error) {
	contentLength, ok := r.Headers.Get(constants.ContentLengthHeader)
	if !ok {
		r.state = Done
		return len(data), nil
	}

	value, err := strconv.Atoi(contentLength)
	if err != nil {
		return 0, err
	}

	r.Body = append(r.Body, data...)
	if len(r.Body) > value {
		return 0, fmt.Errorf("error: body content is greather than pased on content length header")
	}

	if len(r.Body) == value {
		r.state = Done
	}

	return len(data), nil
}
