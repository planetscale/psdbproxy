package psdbproxy

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"time"

	"github.com/golang/glog"
	"github.com/planetscale/psdb/auth"
	"github.com/planetscale/psdb/core/client"
	"vitess.io/vitess/go/mysql"
)

type Server struct {
	Name          string
	Logger        *slog.Logger
	Addr          string
	UpstreamAddr  string
	ReadTimeout   time.Duration
	Authorization *auth.Authorization
	ClientOptions []client.Option

	listener *mysql.Listener

	mysql.UnimplementedHandler
}

func (s *Server) Serve(l net.Listener, authMethod mysql.AuthMethodDescription) error {
	s.ensureSetup()

	handler, err := s.handler()
	if err != nil {
		return err
	}
	if err := handler.testCredentials(5 * time.Second); err != nil {
		return err
	}

	var auth mysql.AuthServer
	switch authMethod {
	case mysql.CachingSha2Password:
		auth = &cachingSha2AuthServerNone{}
	case mysql.MysqlNativePassword:
		auth = mysql.NewAuthServerNone()
	default:
		return fmt.Errorf("unsupported auth method: %v", authMethod)
	}

	listener, err := mysql.NewListenerWithConfig(mysql.ListenerConfig{
		Listener:            l,
		AuthServer:          auth,
		Handler:             handler,
		ConnReadTimeout:     s.ReadTimeout,
		ConnWriteTimeout:    30 * time.Second,
		ConnBufferPooling:   true,
		ConnKeepAlivePeriod: 30 * time.Second,
	})
	if err != nil {
		return err
	}

	s.listener = listener
	listener.Accept()
	return nil
}

func (s *Server) ListenAndServe(authMethod mysql.AuthMethodDescription) error {
	l, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	return s.Serve(l, authMethod)
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
