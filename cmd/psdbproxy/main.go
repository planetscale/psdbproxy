package main

import (
	"log/slog"
	"os"

	"github.com/planetscale/psdb/auth"
	"github.com/planetscale/psdbproxy"
	"github.com/spf13/pflag"
	"vitess.io/vitess/go/mysql"
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
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	s := &psdbproxy.Server{
		Logger:       logger.With(slog.String("component", "proxy")),
		Addr:         *flagListen,
		UpstreamAddr: *flagHost,
		Authorization: auth.NewBasicAuth(
			*flagUsername,
			*flagPassword,
		),
	}

	ch := make(chan error)
	go func() {
		ch <- s.ListenAndServe(mysql.CachingSha2Password)
	}()

	logger.Info(
		"mysql server listening",
		"addr", *flagListen,
	)
	if err := <-ch; err != nil {
		logger.Error(
			"oops",
			"err", err,
		)
	}
}
