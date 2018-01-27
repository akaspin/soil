package api_server

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/supervisor"
	"net/http"
)

type Server struct {
	*supervisor.Control
	log    *logx.Log
	server *http.Server
}

func NewServer(ctx context.Context, log *logx.Log, addr string, router *Router) (s *Server) {
	return &Server{
		Control: supervisor.NewControl(ctx),
		log:     log.GetLog("api", "server"),
		server: &http.Server{
			Addr:    addr,
			Handler: router,
		},
	}
}

func (s *Server) Close() (err error) {
	s.server.Shutdown(s.Ctx())
	s.log.Info("closed")
	return s.Control.Close()
}

func (s *Server) Open() (err error) {
	go func() {
		serveErr := s.server.ListenAndServe()
		if serveErr != nil && serveErr.Error() != "http: Server closed" {
			s.log.Error(serveErr)
			s.Close()
		}
	}()
	s.log.Infof("listening on %s", s.server.Addr)
	return s.Control.Open()
}
