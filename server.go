package mysqlgrpc

import (
	"crypto/tls"
	"io"
	"log/slog"
	"net"
	"time"

	"github.com/golang/glog"
	"github.com/planetscale/psdb/auth"
	"vitess.io/vitess/go/mysql"
)

type Server struct {
	Name          string
	Logger        *slog.Logger
	Addr          string
	UpstreamAddr  string
	Authorization *auth.Authorization
	TLSConfig     *tls.Config
	ServerVersion string

	mysql.UnimplementedHandler
}

func (s *Server) Serve(l net.Listener) error {
	s.ensureSetup()

	handler := s.handler()
	if err := handler.testCredentials(5 * time.Second); err != nil {
		return err
	}

	listener, err := mysql.NewListenerWithConfig(mysql.ListenerConfig{
		Listener:            l,
		AuthServer:          mysql.NewAuthServerNone(),
		Handler:             handler,
		ConnReadTimeout:     30 * time.Second,
		ConnWriteTimeout:    30 * time.Second,
		ConnBufferPooling:   true,
		ConnKeepAlivePeriod: 30 * time.Second,
	})
	if err != nil {
		return err
	}

	if s.TLSConfig != nil {
		listener.TLSConfig.Store(s.TLSConfig.Clone())
		listener.RequireSecureTransport = true
	}

	listener.Accept()
	return nil
}

func (s *Server) ListenAndServe() error {
	l, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	return s.Serve(l)
}

func (s *Server) ensureSetup() {
	if s.Logger == nil {
		s.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	// XXX: suppress all global glog output, since this is internal to vitess
	glog.SetLogger(&glog.LoggerFunc{
		DebugfFunc: func(f string, a ...any) {},
		InfofFunc:  func(f string, a ...any) {},
		WarnfFunc:  func(f string, a ...any) {},
		ErrorfFunc: func(f string, a ...any) {},
	})
}
