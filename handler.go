package psdbproxy

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/planetscale/psdb/core/client"
	psdbpb "github.com/planetscale/psdb/types/psdb/v1alpha1"
	"github.com/planetscale/psdb/types/psdb/v1alpha1/psdbv1alpha1connect"
	querypb "github.com/planetscale/vitess-types/gen/vitess/query/v16"
	"vitess.io/vitess/go/mysql"
	"vitess.io/vitess/go/mysql/replication"
	"vitess.io/vitess/go/mysql/sqlerror"
	"vitess.io/vitess/go/sqltypes"
	vitessquerypb "vitess.io/vitess/go/vt/proto/query"
	"vitess.io/vitess/go/vt/vtenv"
	"vitess.io/vitess/go/vt/vterrors"
)

var errNotImplemented = errors.New("not implemented")

const mysqlVersion = "8.0.34-psdbproxy"

func (s *Server) handler() (*handler, error) {
	env, err := vtenv.New(vtenv.Options{
		MySQLServerVersion: mysqlVersion,
	})
	if err != nil {
		return nil, err
	}

	return &handler{
		logger: s.Logger,
		client: client.New(
			s.UpstreamAddr,
			psdbv1alpha1connect.NewDatabaseClient,
			s.Authorization,
		),
		connections: map[*mysql.Conn]*clientData{},
		env:         env,
	}, nil
}

type handler struct {
	mysql.UnimplementedHandler

	logger *slog.Logger
	client psdbv1alpha1connect.DatabaseClient

	connectionsMu sync.RWMutex
	connections   map[*mysql.Conn]*clientData

	env *vtenv.Environment
}

func (h *handler) testCredentials(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err := h.client.CreateSession(
		ctx, connect.NewRequest(&psdbpb.CreateSessionRequest{}),
	)
	return err
}

func (h *handler) NewConnection(c *mysql.Conn) {
	data := &clientData{
		start:      time.Now(),
		remoteAddr: c.RemoteAddr().String(),
	}
	h.connectionsMu.Lock()
	h.connections[c] = data
	h.connectionsMu.Unlock()

	h.logger.LogAttrs(
		context.Background(),
		slog.LevelDebug,
		"new connection",
		slog.String("addr", data.remoteAddr),
		slog.Int("mysql_id", int(c.ConnectionID)),
	)
}

func (h *handler) ConnectionClosed(c *mysql.Conn) {
	h.connectionsMu.Lock()
	start := h.connections[c].start
	remoteAddr := h.connections[c].remoteAddr
	delete(h.connections, c)
	h.connectionsMu.Unlock()

	h.logger.LogAttrs(
		context.Background(),
		slog.LevelDebug,
		"connection closed",
		slog.String("addr", remoteAddr),
		slog.Int("mysql_id", int(c.ConnectionID)),
		slog.Duration("duration", time.Since(start)),
	)
}

type clientData struct {
	start      time.Time
	remoteAddr string
	Session    *psdbpb.Session
}

func (d *clientData) IsOLAP() bool {
	if d.Session == nil {
		return false
	}
	return d.Session.GetVitessSession().Options.Workload == querypb.ExecuteOptions_OLAP
}

var emptyBindVars = make(map[string]*querypb.BindVariable)

func (h *handler) clientData(c *mysql.Conn) *clientData {
	h.connectionsMu.RLock()
	defer h.connectionsMu.RUnlock()
	return h.connections[c]
}

func (h *handler) ComQuery(c *mysql.Conn, query string, callback func(*sqltypes.Result) error) error {
	data := h.clientData(c)

	defer h.logger.LogAttrs(
		context.Background(),
		slog.LevelDebug,
		"execute",
		slog.String("addr", data.remoteAddr),
		slog.Int("mysql_id", int(c.ConnectionID)),
		slog.String("query", query),
		slog.Bool("olap", data.IsOLAP()),
	)

	if data.IsOLAP() {
		return h.streamExecute(c, data, query, emptyBindVars, callback)
	}

	resp, err := h.client.Execute(context.Background(), connect.NewRequest(&psdbpb.ExecuteRequest{
		Session:       data.Session,
		Query:         query,
		BindVariables: emptyBindVars,
	}))
	if resp != nil && resp.Msg != nil {
		bindSession(c, data, resp.Msg.GetSession())
	}
	if err != nil {
		return sqlerror.NewSQLErrorFromError(err)
	}
	if resp.Msg.Error != nil {
		return sqlerror.NewSQLErrorFromError(vterrors.FromVTRPC(
			castRPCError(resp.Msg.Error)),
		)
	}

	return callback(sqltypes.Proto3ToResult(
		castQueryResult(resp.Msg.GetResult())),
	)
}

func (h *handler) ComPrepare(c *mysql.Conn, query string, bindVars map[string]*vitessquerypb.BindVariable) ([]*vitessquerypb.Field, error) {
	data := h.clientData(c)

	defer h.logger.LogAttrs(
		context.Background(),
		slog.LevelDebug,
		"prepare",
		slog.String("addr", data.remoteAddr),
		slog.Int("mysql_id", int(c.ConnectionID)),
		slog.String("query", query),
	)

	resp, err := h.client.Prepare(context.Background(), connect.NewRequest(&psdbpb.PrepareRequest{
		Session:       data.Session,
		Query:         query,
		BindVariables: castBindVars(bindVars),
	}))

	var out bytes.Buffer

	out.Write([]byte(query))
	out.Write([]byte("\n"))
	json.NewEncoder(&out).Encode(resp)
	io.Copy(os.Stdout, &out)

	if resp != nil && resp.Msg != nil {
		bindSession(c, data, resp.Msg.GetSession())
	}
	if err != nil {
		return nil, sqlerror.NewSQLErrorFromError(err)
	}
	if resp.Msg.Error != nil {
		return nil, sqlerror.NewSQLErrorFromError(vterrors.FromVTRPC(
			castRPCError(resp.Msg.Error)),
		)
	}

	return castFields(resp.Msg.GetFields()), nil
}

func (h *handler) ComStmtExecute(c *mysql.Conn, prepare *mysql.PrepareData, callback func(*sqltypes.Result) error) error {
	data := h.clientData(c)

	defer h.logger.LogAttrs(
		context.Background(),
		slog.LevelDebug,
		"stmt_execute",
		slog.String("addr", data.remoteAddr),
		slog.Int("mysql_id", int(c.ConnectionID)),
		slog.String("query", prepare.PrepareStmt),
		slog.Bool("olap", data.IsOLAP()),
	)

	if data.IsOLAP() {
		return h.streamExecute(c, data, prepare.PrepareStmt, castBindVars(prepare.BindVars), callback)
	}

	resp, err := h.client.Execute(context.Background(), connect.NewRequest(&psdbpb.ExecuteRequest{
		Session:       data.Session,
		Query:         prepare.PrepareStmt,
		BindVariables: castBindVars(prepare.BindVars),
	}))
	if resp != nil && resp.Msg != nil {
		bindSession(c, data, resp.Msg.GetSession())
	}
	if err != nil {
		return sqlerror.NewSQLErrorFromError(err)
	}
	if resp.Msg.Error != nil {
		return sqlerror.NewSQLErrorFromError(vterrors.FromVTRPC(
			castRPCError(resp.Msg.Error)),
		)
	}

	return callback(sqltypes.Proto3ToResult(
		castQueryResult(resp.Msg.GetResult())),
	)
}

func (h *handler) ComRegisterReplica(c *mysql.Conn, replicaHost string, replicaPort uint16, replicaUser string, replicaPassword string) error {
	return errNotImplemented
}

func (h *handler) ComBinlogDump(c *mysql.Conn, logFile string, binlogPos uint32) error {
	return errNotImplemented
}

func (h *handler) ComBinlogDumpGTID(c *mysql.Conn, logFile string, logPos uint64, gtidSet replication.GTIDSet) error {
	return errNotImplemented
}

func (h *handler) WarningCount(c *mysql.Conn) uint16 {
	session := h.clientData(c).Session
	if session == nil {
		return 0
	}
	return uint16(len(session.GetVitessSession().GetWarnings()))
}

func (h *handler) streamExecute(c *mysql.Conn, data *clientData, query string, bindVars map[string]*querypb.BindVariable, callback func(*sqltypes.Result) error) error {
	stream, err := h.client.StreamExecute(context.Background(), connect.NewRequest(&psdbpb.ExecuteRequest{
		Session:       data.Session,
		Query:         query,
		BindVariables: bindVars,
	}))
	if err != nil {
		return sqlerror.NewSQLErrorFromError(err)
	}

	var fields []*querypb.Field
	var resp *psdbpb.ExecuteResponse

	for stream.Receive() {
		resp = stream.Msg()
		// NOTE: Some results do not have any Result. This is most likely
		// the case when a Session is returned. While Vitess currently (as of v18)
		// is implemented such that the last streaming response
		// contains a Session, but not Result, I do not want to assume
		// this is always the case, so this is implemented to handle
		// both existing or none existing, or either existing to cover
		// our bases.

		// Some results may contain a Session, if so
		// we need to bind it to the mysql.Conn like normal
		if resp.Session != nil {
			bindSession(c, data, resp.Session)

			// XXX: bindSession copies the pointer into data.Session
			// so resp.Session here cannot be mutated otherwise we'll
			// mutate the Session stored inside of data.
			// Specifically doing a resp.Session.Reset() will
			// zero out the struct that is being referenced now
			// in data.session.
			resp.Session = nil
		}

		// If we have ane error, we just return the error
		if resp.Error != nil {
			return sqlerror.NewSQLErrorFromError(vterrors.FromVTRPC(
				castRPCError(resp.Error)),
			)
		}

		// Lastly if there are results, we return them to the mysql client.
		// messages without results get ignored at this point since they
		// likely only contained session data.
		if resp.Result != nil {
			if fields == nil {
				fields = resp.Result.GetFields()
			}
			if err := callback(sqltypes.CustomProto3ToResult(
				castFields(fields), castQueryResult(resp.GetResult())),
			); err != nil {
				return err
			}
		}

		// For each iteration, stream.Msg() is reused to the same struct,
		// We can do resp.Reset(). This nulls out everything, but doesn't let
		// anything me reused. This is nuclear, but it would get the job done.
		// Doing this means `resp.Result` gets nulled, and needs to be GC'd for each page.
		// hitting each field explicitly with a `Reset()` allows us to reuse memory
		// safely and efficiently between.
		// See: https://pkg.go.dev/google.golang.org/protobuf/proto#Reset
		// So this maintains the actual structs, but zeroed out.

		if resp.Result != nil {
			resp.Result.Reset()
		}
		if resp.Error != nil {
			resp.Error.Reset()
		}
	}

	if err := stream.Err(); err != nil {
		return sqlerror.NewSQLErrorFromError(err)
	}

	return nil
}

func (h *handler) Env() *vtenv.Environment {
	return h.env
}

func bindSession(c *mysql.Conn, data *clientData, session *psdbpb.Session) {
	if session == nil || session.VitessSession == nil {
		return
	}

	if session.VitessSession.InTransaction {
		c.StatusFlags |= mysql.ServerStatusInTrans
	} else {
		c.StatusFlags &= mysql.NoServerStatusInTrans
	}
	if session.VitessSession.Autocommit {
		c.StatusFlags |= mysql.ServerStatusAutocommit
	} else {
		c.StatusFlags &= mysql.NoServerStatusAutocommit
	}

	data.Session = session
}
