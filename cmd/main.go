package main

import (
	"flag"
	"os"

	"github.com/planetscale/mysqlgrpc"
	"github.com/planetscale/psdb/auth"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// separate entirely separate flagset from vitess
	commandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)

	flagListen   = commandLine.StringP("listen", "l", "127.0.0.1:3306", "mysql address to listen")
	flagHost     = commandLine.StringP("host", "h", "aws.connect.psdb.cloud", "upstream PlanetScale hostname")
	flagUsername = commandLine.StringP("username", "u", "", "PlanetScale username")
	flagPassword = commandLine.StringP("password", "p", "", "PlanetScale password")
)

func init() {
	commandLine.Parse(os.Args[1:])

	// needed to make vitess happy
	flag.CommandLine.Parse([]string{})
}

func makeLogger(level zapcore.Level) *zap.Logger {
	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(level)
	l, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	return l
}

func main() {
	logger := makeLogger(zap.DebugLevel)
	defer logger.Sync()

	s := &mysqlgrpc.Server{
		Logger:       logger.With(zap.String("component", "proxy")),
		Addr:         *flagListen,
		UpstreamAddr: *flagHost,
		Authorization: auth.NewBasicAuth(
			*flagUsername,
			*flagPassword,
		),
	}

	ch := make(chan error)
	go func() {
		ch <- s.ListenAndServe()
	}()

	logger.Info("mysql server listening", zap.String("addr", *flagListen))
	if err := <-ch; err != nil {
		logger.Fatal("oops", zap.Error(err))
	}
}
