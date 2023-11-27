package mysqlgrpc

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
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

	listener *mysql.Listener

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

	s.listener = listener
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

func (s *Server) Shutdown() {
	if s.listener != nil {
		s.listener.Shutdown()
	}
}

func (s *Server) Close() {
	if s.listener != nil {
		s.listener.Close()
	}
}

func (s *Server) ensureSetup() {
	if s.Logger == nil {
		s.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	// XXX: vitess requires default commandline flags to be parsed,
	// but if they are not, fake it so it doesn't panic.
	if !flag.Parsed() {
		flag.CommandLine.Parse([]string{})
	}

	vtLogger := s.Logger.With("component", "vitess")

	vtLog := func(f string, a ...any) {
		if vtLogger.Enabled(context.Background(), slog.LevelDebug) {
			vtLogger.Debug(fmt.Sprintf(f, a...))
		}
	}

	// XXX: suppress all global glog output, since this is internal to vitess
	glog.SetLogger(&glog.LoggerFunc{
		DebugfFunc: vtLog,
		InfofFunc:  vtLog,
		WarnfFunc:  vtLog,
		ErrorfFunc: vtLog,
	})
}
