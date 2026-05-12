package api

import "fmt"

func (s *Server) publicStatusURL(slug string) string {
	if s.Config.Env == "dev" {
		return fmt.Sprintf("http://localhost:%s/status/%s", s.Config.Port, slug)
	}
	return fmt.Sprintf("https://%s/status/%s", s.Config.Domain, slug)
}

func (s *Server) publicHomeURL() string {
	if s.Config.Env == "dev" {
		return fmt.Sprintf("http://localhost:%s", s.Config.Port)
	}
	return fmt.Sprintf("https://%s", s.Config.Domain)
}
