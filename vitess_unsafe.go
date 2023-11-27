package psdbproxy

// The unsafe code in this file is a 0 allocation type cast. This is unsafe if the types
// are not actually equivalent. Their memory layouts must be identical, and in our case,
// all of these types are exactly the same, just come from a different package.
// The alternative is to do a manual allocation and copy through a deeply nested structure
// that is both expensive in terms of CPU and memory, as well as writing and maintaining
// how to transform every single type.
// Hopefully at some point, these `vitess.io/vitess/go/vt` types aren't needed at all, but right
// now they're only needed for the boundary for passing structs to the embedded vtgate mysql
// server since the embedded server expects _their_ types, and our code expects ours. All of
// our RPC code uses our types to not have a dependency on vitess as a whole.

import (
	"unsafe"

	querypb "github.com/planetscale/vitess-types/gen/vitess/query/v16"
	vtrpcpb "github.com/planetscale/vitess-types/gen/vitess/vtrpc/v16"
	vitessquerypb "vitess.io/vitess/go/vt/proto/query"
	vitessvtrpcpb "vitess.io/vitess/go/vt/proto/vtrpc"
)

// castTo likely shouldn't be used directly and is the most unsafe. It does 0 type checking,
// or compatibility checking, and is the lowest form of "cast this type to this type" as you might
// do in C. If you need to cast other types, please add a new "cast*" function like below that
// is for each type explicitly.
func castTo[RT any, T any](a T) RT {
	return (*(*RT)(unsafe.Pointer(&a)))
}

func castBindVars(bv map[string]*vitessquerypb.BindVariable) map[string]*querypb.BindVariable {
	return castTo[map[string]*querypb.BindVariable](bv)
}

func castRPCError(e *vtrpcpb.RPCError) *vitessvtrpcpb.RPCError {
	return castTo[*vitessvtrpcpb.RPCError](e)
}

func castFields(fields []*querypb.Field) []*vitessquerypb.Field {
	return castTo[[]*vitessquerypb.Field](fields)
}

func castQueryResult(qr *querypb.QueryResult) *vitessquerypb.QueryResult {
	return castTo[*vitessquerypb.QueryResult](qr)
}
