package psdbproxy

import (
	"crypto/rand"
	"net"

	"vitess.io/vitess/go/mysql"
	querypb "vitess.io/vitess/go/vt/proto/query"
)

// cachingSha2AuthServerNone takes all comers.
type cachingSha2AuthServerNone struct{}

type noneGetter struct{}

// AuthMethods returns the list of registered auth methods
// implemented by this auth server.
func (a *cachingSha2AuthServerNone) AuthMethods() []mysql.AuthMethod {
	return []mysql.AuthMethod{&mysqlCachingSha2AuthMethod{}}
}

// DefaultAuthMethodDescription returns MysqlNativePassword as the default
// authentication method for the auth server implementation.
func (a *cachingSha2AuthServerNone) DefaultAuthMethodDescription() mysql.AuthMethodDescription {
	return mysql.CachingSha2Password
}

// Get returns the empty string
func (ng *noneGetter) Get() *querypb.VTGateCallerID {
	return &querypb.VTGateCallerID{Username: "root"}
}

type mysqlCachingSha2AuthMethod struct{}

func (n *mysqlCachingSha2AuthMethod) Name() mysql.AuthMethodDescription {
	return mysql.CachingSha2Password
}

func (n *mysqlCachingSha2AuthMethod) HandleUser(conn *mysql.Conn, user string) bool {
	return true
}

func (n *mysqlCachingSha2AuthMethod) AuthPluginData() ([]byte, error) {
	salt, err := newSalt()
	if err != nil {
		return nil, err
	}
	return append(salt, 0), nil
}

func (n *mysqlCachingSha2AuthMethod) AllowClearTextWithoutTLS() bool {
	return true
}

func (n *mysqlCachingSha2AuthMethod) HandleAuthPluginData(c *mysql.Conn, user string, serverAuthPluginData []byte, clientAuthPluginData []byte, remoteAddr net.Addr) (mysql.Getter, error) {
	return &noneGetter{}, nil
}

func newSalt() ([]byte, error) {
	salt := make([]byte, 20)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}

	// Salt must be a legal UTF8 string.
	for i := range len(salt) {
		salt[i] &= 0x7f
		if salt[i] == '\x00' || salt[i] == '$' {
			salt[i]++
		}
	}

	return salt, nil
}
