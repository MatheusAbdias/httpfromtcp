package response

import (
	"fmt"
	"httpfromtcp/internal/constants"
	"httpfromtcp/internal/headers"
	"io"
	"strconv"
)

type StatusCode int

const (
	OK                  StatusCode = 200
	BadRequest          StatusCode = 400
	InternalServerError StatusCode = 500
)

type writerState int

const (
	writeStatusLine writerState = iota
	writeHeaders
	writeBody
	writeTrailers
)

type Writer struct {
	buf   io.Writer
	state writerState
}

func NewResponseWriter(buf io.Writer) *Writer {
	return &Writer{
		buf:   buf,
		state: writeStatusLine,
	}
}

func (w *Writer) write(data []byte) (int, error) {
	return w.buf.Write(data)
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.state != writeStatusLine {
		return fmt.Errorf("Invalid status: %v", w.state)
	}

	switch statusCode {
	case OK:
		_, err := w.write([]byte("HTTP/1.1 200 OK\r\n"))
		if err != nil {
			return err
		}

	case BadRequest:
		_, err := w.write([]byte("HTTP/1.1 400 Bad Request\r\n"))
		if err != nil {
			return err
		}

	case InternalServerError:
		_, err := w.write([]byte("HTTP/1.1 500 Internal Server Error\r\n"))
		if err != nil {
			return err
		}
	}

	w.state = writeHeaders
	return nil
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.state != writeHeaders {
		return fmt.Errorf("Invalid status: %v", w.state)
	}

	for key, value := range headers {
		fieldLine := fmt.Sprintf("%s: %s\r\n", key, value)
		_, err := w.write([]byte(fieldLine))
		if err != nil {
			return err
		}
	}

	_, err := w.write([]byte("\r\n"))
	if err != nil {
		return err
	}

	w.state = writeBody
	return nil
}

func (w *Writer) WriteBody(data []byte) (int, error) {
	if w.state != writeBody {
		return 0, fmt.Errorf("Invalid status: %v", w.state)
	}

	return w.write(data)
}

func (w *Writer) WriteChunkedBody(data []byte) (int, error) {
	if w.state != writeBody {
		return 0, fmt.Errorf("Invalid status: %v", w.state)
	}

	chunkHeader := fmt.Sprintf("%x\r\n", len(data))
	data = append([]byte(chunkHeader), data...)
	data = append(data, []byte("\r\n")...)

	return w.write(data)
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.state != writeBody {
		return 0, fmt.Errorf("Invalid status: %v", w.state)
	}

	w.state = writeTrailers
	return w.write([]byte("0\r\n"))
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	return headers.Headers{
		constants.ContentLengthHeader: strconv.Itoa(contentLen),
		"connection":                  "close",
		"content-type":                "text/plain",
	}
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	if w.state != writeTrailers {
		return fmt.Errorf("Invalid Status: %v", w.state)
	}

	for key, value := range h {
		fieldLine := fmt.Sprintf("%s: %s\r\n", key, value)
		_, err := w.write([]byte(fieldLine))
		if err != nil {
			return err
		}
	}

	_, err := w.write([]byte("\r\n"))
	if err != nil {
		return err
	}

	return nil
}
