package server

import (
	"fmt"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"log"
	"net"
	"sync/atomic"
)

type Server struct {
	ln       net.Listener
	handler  Handler
	isClosed atomic.Bool
}

func (s *Server) Close() error {
	if s.isClosed.Swap(true) {
		return nil
	}

	if s.ln != nil {
		return s.ln.Close()
	}
	return nil
}

func Serve(port int, handler Handler) (*Server, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		return nil, err
	}

	server := &Server{ln: ln, handler: handler}

	go server.listen()

	return server, nil
}

func (s *Server) listen() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			if s.isClosed.Load() {
				return
			}
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	resp := response.NewResponseWriter(conn)
	req, err := request.RequestFromReader(conn)
	if err != nil {
		panic(err)
	}

	s.handler(req, resp)
}
