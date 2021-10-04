package gossm

import (
	"fmt"
)

type Servers []*Server

type Server struct {
	Name          string `json:"name"`
	URL           string `json:"url"`
	IPAddress     string `json:"ipAddress"`
	Port          int    `json:"port"`
	Protocol      string `json:"protocol"`
	CheckInterval int    `json:"checkInterval"`
	Timeout       int    `json:"timeout"`
}

func (s *Server) String() string {
	return fmt.Sprintf("%s %s:%d", s.Protocol, s.IPAddress, s.Port)
}

func (s *Server) MarshalText() (text []byte, err error) {
	return []byte(s.String()), nil
}

func (s *Server) FormattedAddress() string {
	if s.Protocol == "http" {
		return s.URL
	}
	return fmt.Sprintf("%s:%d", s.IPAddress, s.Port)
}
