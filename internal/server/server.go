package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"
)

// type Server struct - Contains the state of the server
// func Serve(port int) (*Server, error) - Creates a net.Listener and returns a new Server instance. Starts listening for requests inside a goroutine.
// func (s *Server) Close() error - Closes the listener and the server
// func (s *Server) listen() - Uses a loop to .Accept new connections as they come in, and handles each one in a new goroutine. I used an atomic.Bool to track whether the server is closed or not so that I can ignore connection errors after the server is closed.
// func (s *Server) handle(conn net.Conn) - Handles a single connection by writing the following response and then closing the connection:

type Server struct {
	closed   atomic.Bool
	listener net.Listener
}

func Serve(port int) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))

	if err != nil {
		return nil, err
	}

	server := &Server{
		listener: listener,
	}

	go server.listen()

	return server, nil
}

func (s *Server) Close() error {
	s.closed.Store(true)
	if s.listener != nil {
		return s.listener.Close()
	}

	return nil
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			log.Printf("Error accepting a connection: %v", err)
			continue
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	response := "HTTP/1.1 200 OK\r\n" +
		"Content-Type: text/plain\r\n" +
		"\r\n" +
		"Hello World!\n"

	conn.Write([]byte(response))

	return
}
