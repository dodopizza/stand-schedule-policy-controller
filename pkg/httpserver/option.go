package httpserver

import (
	"net"
	"strconv"
)

type (
	Option func(*Server)
)

func Port(port int) Option {
	return func(s *Server) {
		s.server.Addr = net.JoinHostPort("", strconv.Itoa(port))
	}
}
