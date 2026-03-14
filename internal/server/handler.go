package server

import (
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
)

type Handler func(req *request.Request, resp *response.Writer)

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

func (h *HandlerError) Write(resp *response.Writer) {
	resp.WriteStatusLine(h.StatusCode)
	resp.WriteHeaders(response.GetDefaultHeaders(len(h.Message)))
	resp.WriteBody([]byte(h.Message))
}
