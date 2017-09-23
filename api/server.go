package api

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/supervisor"
	"net/http"
)

type Server struct {
	*supervisor.Control
	trap *supervisor.Trap
	log  *logx.Log

	server *http.Server
}

func NewServer(ctx context.Context, log *logx.Log, addr string, router *Router) (s *Server) {
	mux := http.NewServeMux()
	s = &Server{
		Control: supervisor.NewControl(ctx),
		log:     log.GetLog("api", "server"),
		server: &http.Server{
			Addr:    addr,
			Handler: mux,
		},
	}
	s.trap = supervisor.NewTrap(s.Control.Cancel)
	router.Bind(ctx, s.log, mux)
	return
}

func (s *Server) Close() (err error) {
	s.log.Debug("closing")
	s.server.Shutdown(s.Ctx())
	err = s.Control.Close()
	s.log.Info("close")
	return
}

func (s *Server) Open() (err error) {
	s.Acquire()
	go func() {
		defer s.Release()
		serveErr := s.server.ListenAndServe()
		if serveErr != nil && serveErr.Error() != "http: Server closed" {
			s.log.Error(serveErr)
			s.trap.Catch(serveErr)
		}
	}()
	err = s.Control.Open()
	s.log.Infof("listening on %s", s.server.Addr)
	return
}
