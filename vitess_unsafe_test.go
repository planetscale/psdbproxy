package psdbproxy

import (
	"maps"
	"reflect"
	"slices"
	"strings"
	"testing"
	"unsafe"

	querypb "github.com/planetscale/vitess-types/gen/vitess/query/v21"
	vtrpcpb "github.com/planetscale/vitess-types/gen/vitess/vtrpc/v21"
	vitessquerypb "vitess.io/vitess/go/vt/proto/query"
	vitessvtrpcpb "vitess.io/vitess/go/vt/proto/vtrpc"
)

// TestCastLayoutCompatibility guards the invariant the casts in vitess_unsafe.go
// rely on: each pair of types must have an identical memory layout. If a
// dependency update regenerates the protobuf types with a different field
// layout, this test fails with a clear diagnostic instead of the data-corruption
// failures the tests below would otherwise report.
func TestCastLayoutCompatibility(t *testing.T) {
	check := func(t *testing.T, ours, vitess any) {
		t.Helper()
		ot, vt := reflect.TypeOf(ours).Elem(), reflect.TypeOf(vitess).Elem()
		if unsafe.Sizeof(ours) != unsafe.Sizeof(vitess) {
			t.Errorf("%s: size mismatch: ours=%d vitess=%d", ot.Name(), unsafe.Sizeof(ours), unsafe.Sizeof(vitess))
		}
		for i := 0; i < ot.NumField(); i++ {
			of := ot.Field(i)
			if !of.IsExported() {
				continue
			}
			vf, ok := vt.FieldByName(of.Name)
			if !ok {
				t.Errorf("%s: vitess type %s is missing field %s", ot.Name(), vt.Name(), of.Name)
				continue
			}
			if of.Offset != vf.Offset {
				t.Errorf("%s.%s: offset mismatch: ours=%d vitess=%d", ot.Name(), of.Name, of.Offset, vf.Offset)
			}
		}
	}

	t.Run("BindVariable", func(t *testing.T) { check(t, &querypb.BindVariable{}, &vitessquerypb.BindVariable{}) })
	t.Run("Value", func(t *testing.T) { check(t, &querypb.Value{}, &vitessquerypb.Value{}) })
	t.Run("Field", func(t *testing.T) { check(t, &querypb.Field{}, &vitessquerypb.Field{}) })
	t.Run("Row", func(t *testing.T) { check(t, &querypb.Row{}, &vitessquerypb.Row{}) })
	t.Run("QueryResult", func(t *testing.T) { check(t, &querypb.QueryResult{}, &vitessquerypb.QueryResult{}) })
	t.Run("RPCError", func(t *testing.T) { check(t, &vtrpcpb.RPCError{}, &vitessvtrpcpb.RPCError{}) })
}

func TestCasts(t *testing.T) {
	t.Run("castBindVars", func(t *testing.T) {
		in := map[string]*vitessquerypb.BindVariable{
			"int": {
				Type:  vitessquerypb.Type_INT64,
				Value: []byte("42"),
			},
			"tuple": {
				Type: vitessquerypb.Type_TUPLE,
				Values: []*vitessquerypb.Value{
					{Type: vitessquerypb.Type_VARCHAR, Value: []byte("foo")},
					{Type: vitessquerypb.Type_INT64, Value: []byte("7")},
				},
			},
		}

		out := castBindVars(in)

		outKeys := slices.Collect(maps.Keys(out))
		slices.Sort(outKeys)
		if !slices.Equal([]string{"int", "tuple"}, outKeys) {
			t.Errorf("bind var keys don't match\nwant: int, tuple\ngot:  %s", strings.Join(outKeys, ", "))
		}

		if got := out["int"]; got == nil {
			t.Errorf("missing 'int'")
		} else {
			if got.Type != querypb.Type_INT64 {
				t.Errorf("int.type: want Type_INT64 (%v), got %v", querypb.Type_INT64, got.Type)
			}
			if string(got.Value) != "42" {
				t.Errorf("int.value: want '42', got '%s'", got.Value)
			}
			if len(got.Values) != 0 {
				t.Errorf("int.values: want none, got %v", got.Values)
			}
		}

		if tuple := out["tuple"]; tuple == nil {
			t.Errorf("missing 'tuple'")
		} else {
			if tuple.Type != querypb.Type_TUPLE {
				t.Errorf("tuple.type: want Type_TUPLE (%v), got %v", querypb.Type_TUPLE, tuple.Type)
			}
			if len(tuple.Value) != 0 {
				t.Errorf("tuple.value: want nothing, got %v", tuple.Value)
			}
			if len(tuple.Values) != 2 {
				t.Errorf("tuple.values: want 2 items, got %d", len(tuple.Values))
			}
			if len(tuple.Values) >= 1 {
				if tuple.Values[0].Type != querypb.Type_VARCHAR {
					t.Errorf("tuple.values[0].type: want Type_VARCHAR (%v), got %v", querypb.Type_VARCHAR, tuple.Values[0].Type)
				}
				if string(tuple.Values[0].Value) != "foo" {
					t.Errorf("tuple.values[0].value: want 'foo', got %q", tuple.Values[0].Value)
				}
			}
			if len(tuple.Values) >= 2 {
				if tuple.Values[1].Type != querypb.Type_INT64 {
					t.Errorf("tuple.values[1]: want Type_INT64 (%v), got %v", querypb.Type_INT64, tuple.Values[1].Type)
				}
				if string(tuple.Values[1].Value) != "7" {
					t.Errorf("tuple.values[0].value: want '7', got %q", tuple.Values[1].Value)
				}
			}
		}

		// the cast is zero-copy: mutations through the cast map must be
		// visible through the original map and vice versa.
		out["int"].Value = []byte("43")
		if string(in["int"].Value) != "43" {
			t.Errorf("cast copied the map entries: mutation through cast not visible on original")
		}
		in["extra"] = &vitessquerypb.BindVariable{Type: vitessquerypb.Type_VARCHAR, Value: []byte("bar")}
		if got := out["extra"]; got == nil || string(got.Value) != "bar" {
			t.Errorf("cast copied the map: new entry not visible through cast")
		}
	})

	t.Run("castRPCError", func(t *testing.T) {
		in := &vtrpcpb.RPCError{
			Message: "syntax error",
			Code:    vtrpcpb.Code_INVALID_ARGUMENT,
		}

		out := castRPCError(in)

		if out.Message != in.Message {
			t.Errorf("expected message %q, got %q", in.Message, out.Message)
		}
		if out.Code != vitessvtrpcpb.Code(in.Code) {
			t.Errorf("expected code %v, got %v", in.Code, out.Code)
		}

		// the cast is zero-copy: it must return the same pointer.
		out.Message = "other error"
		if in.Message != "other error" {
			t.Errorf("cast copied the RPCError: mutation through cast not visible on original")
		}
	})

	t.Run("castFields", func(t *testing.T) {
		in := []*querypb.Field{
			{
				Name:         "id",
				Type:         querypb.Type_INT64,
				Table:        "users",
				OrgTable:     "users",
				Database:     "main",
				OrgName:      "id",
				ColumnLength: 11,
				Charset:      63,
				Decimals:     0,
				Flags:        0x1021,
				ColumnType:   "bigint(20)",
			},
			{
				Name: "name",
				Type: querypb.Type_VARCHAR,
			},
		}

		out := castFields(in)

		if len(out) != len(in) {
			t.Fatalf("expected %d fields, got %d", len(in), len(out))
		}
		f := out[0]
		if f.Name != "id" || f.Type != vitessquerypb.Type_INT64 ||
			f.Table != "users" || f.OrgTable != "users" || f.Database != "main" ||
			f.OrgName != "id" || f.ColumnLength != 11 || f.Charset != 63 ||
			f.Decimals != 0 || f.Flags != 0x1021 || f.ColumnType != "bigint(20)" {
			t.Errorf("cast lost data for field 0: %+v", f)
		}
		if f := out[1]; f.Name != "name" || f.Type != vitessquerypb.Type_VARCHAR {
			t.Errorf("cast lost data for field 1: %+v", f)
		}

		// the cast is zero-copy: mutations through the cast slice must be
		// visible through the original slice and vice versa.
		out[0].Name = "user_id"
		if in[0].Name != "user_id" {
			t.Errorf("cast copied the fields: mutation through cast not visible on original")
		}
		in[1].Name = "username"
		if out[1].Name != "username" {
			t.Errorf("cast copied the fields: mutation on original not visible through cast")
		}
	})

	t.Run("castQueryResult", func(t *testing.T) {
		in := &querypb.QueryResult{
			Fields: []*querypb.Field{
				{Name: "id", Type: querypb.Type_INT64},
				{Name: "name", Type: querypb.Type_VARCHAR},
			},
			RowsAffected: 2,
			InsertId:     42,
			Rows: []*querypb.Row{
				{Lengths: []int64{1, 5}, Values: []byte("1alice")},
				{Lengths: []int64{2, -1}, Values: []byte("42")},
			},
			Info:                "ok",
			SessionStateChanges: "state",
		}

		out := castQueryResult(in)

		if out.RowsAffected != in.RowsAffected {
			t.Errorf("expected RowsAffected %d, got %d", in.RowsAffected, out.RowsAffected)
		}
		if out.InsertId != in.InsertId {
			t.Errorf("expected InsertId %d, got %d", in.InsertId, out.InsertId)
		}
		if out.Info != in.Info {
			t.Errorf("expected Info %q, got %q", in.Info, out.Info)
		}
		if out.SessionStateChanges != in.SessionStateChanges {
			t.Errorf("expected SessionStateChanges %q, got %q", in.SessionStateChanges, out.SessionStateChanges)
		}
		if len(out.Fields) != len(in.Fields) {
			t.Fatalf("expected %d fields, got %d", len(in.Fields), len(out.Fields))
		}
		if out.Fields[0].Name != "id" || out.Fields[0].Type != vitessquerypb.Type_INT64 {
			t.Errorf("cast lost data for field 0: %+v", out.Fields[0])
		}
		if len(out.Rows) != len(in.Rows) {
			t.Fatalf("expected %d rows, got %d", len(in.Rows), len(out.Rows))
		}
		if got := out.Rows[0]; len(got.Lengths) != 2 || got.Lengths[0] != 1 || got.Lengths[1] != 5 || string(got.Values) != "1alice" {
			t.Errorf("cast lost data for row 0: %+v", got)
		}
		if got := out.Rows[1]; len(got.Lengths) != 2 || got.Lengths[0] != 2 || got.Lengths[1] != -1 || string(got.Values) != "42" {
			t.Errorf("cast lost data for row 1: %+v", got)
		}

		// the cast is zero-copy: it must return the same pointer.
		out.Info = "updated"
		if in.Info != "updated" {
			t.Errorf("cast copied the QueryResult: mutation through cast not visible on original")
		}
	})
}
