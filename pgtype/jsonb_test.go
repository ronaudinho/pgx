package pgtype_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/pgtype/testutil"
)

func TestJsonbTranscode(t *testing.T) {
	conn := testutil.MustConnectPgx(t)
	defer testutil.MustClose(t, conn)
	if _, ok := conn.ConnInfo.DataTypeForName("jsonb"); !ok {
		t.Skip("Skipping due to no jsonb type")
	}

	testutil.TestSuccessfulTranscode(t, "jsonb", []interface{}{
		pgtype.Jsonb{Bytes: []byte("{}"), Status: pgtype.Present},
		pgtype.Jsonb{Bytes: []byte("null"), Status: pgtype.Present},
		pgtype.Jsonb{Bytes: []byte("42"), Status: pgtype.Present},
		pgtype.Jsonb{Bytes: []byte(`"hello"`), Status: pgtype.Present},
		pgtype.Jsonb{Status: pgtype.Null},
	})
}

func TestJsonbSet(t *testing.T) {
	successfulTests := []struct {
		source interface{}
		result pgtype.Jsonb
	}{
		{source: "{}", result: pgtype.Jsonb{Bytes: []byte("{}"), Status: pgtype.Present}},
		{source: []byte("{}"), result: pgtype.Jsonb{Bytes: []byte("{}"), Status: pgtype.Present}},
		{source: ([]byte)(nil), result: pgtype.Jsonb{Status: pgtype.Null}},
		{source: (*string)(nil), result: pgtype.Jsonb{Status: pgtype.Null}},
		{source: []int{1, 2, 3}, result: pgtype.Jsonb{Bytes: []byte("[1,2,3]"), Status: pgtype.Present}},
		{source: map[string]interface{}{"foo": "bar"}, result: pgtype.Jsonb{Bytes: []byte(`{"foo":"bar"}`), Status: pgtype.Present}},
	}

	for i, tt := range successfulTests {
		var d pgtype.Jsonb
		err := d.Set(tt.source)
		if err != nil {
			t.Errorf("%d: %v", i, err)
		}

		if !reflect.DeepEqual(d, tt.result) {
			t.Errorf("%d: expected %v to convert to %v, but it was %v", i, tt.source, tt.result, d)
		}
	}
}

func TestJsonbAssignTo(t *testing.T) {
	var s string
	var ps *string
	var b []byte

	rawStringTests := []struct {
		src      pgtype.Jsonb
		dst      *string
		expected string
	}{
		{src: pgtype.Jsonb{Bytes: []byte("{}"), Status: pgtype.Present}, dst: &s, expected: "{}"},
	}

	for i, tt := range rawStringTests {
		err := tt.src.AssignTo(tt.dst)
		if err != nil {
			t.Errorf("%d: %v", i, err)
		}

		if *tt.dst != tt.expected {
			t.Errorf("%d: expected %v to assign %v, but result was %v", i, tt.src, tt.expected, *tt.dst)
		}
	}

	rawBytesTests := []struct {
		src      pgtype.Jsonb
		dst      *[]byte
		expected []byte
	}{
		{src: pgtype.Jsonb{Bytes: []byte("{}"), Status: pgtype.Present}, dst: &b, expected: []byte("{}")},
		{src: pgtype.Jsonb{Status: pgtype.Null}, dst: &b, expected: (([]byte)(nil))},
	}

	for i, tt := range rawBytesTests {
		err := tt.src.AssignTo(tt.dst)
		if err != nil {
			t.Errorf("%d: %v", i, err)
		}

		if bytes.Compare(tt.expected, *tt.dst) != 0 {
			t.Errorf("%d: expected %v to assign %v, but result was %v", i, tt.src, tt.expected, *tt.dst)
		}
	}

	var mapDst map[string]interface{}
	type structDst struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	var strDst structDst

	unmarshalTests := []struct {
		src      pgtype.Jsonb
		dst      interface{}
		expected interface{}
	}{
		{src: pgtype.Jsonb{Bytes: []byte(`{"foo":"bar"}`), Status: pgtype.Present}, dst: &mapDst, expected: map[string]interface{}{"foo": "bar"}},
		{src: pgtype.Jsonb{Bytes: []byte(`{"name":"John","age":42}`), Status: pgtype.Present}, dst: &strDst, expected: structDst{Name: "John", Age: 42}},
	}
	for i, tt := range unmarshalTests {
		err := tt.src.AssignTo(tt.dst)
		if err != nil {
			t.Errorf("%d: %v", i, err)
		}

		if dst := reflect.ValueOf(tt.dst).Elem().Interface(); !reflect.DeepEqual(dst, tt.expected) {
			t.Errorf("%d: expected %v to assign %v, but result was %v", i, tt.src, tt.expected, dst)
		}
	}

	pointerAllocTests := []struct {
		src      pgtype.Jsonb
		dst      **string
		expected *string
	}{
		{src: pgtype.Jsonb{Status: pgtype.Null}, dst: &ps, expected: ((*string)(nil))},
	}

	for i, tt := range pointerAllocTests {
		err := tt.src.AssignTo(tt.dst)
		if err != nil {
			t.Errorf("%d: %v", i, err)
		}

		if *tt.dst == tt.expected {
			t.Errorf("%d: expected %v to assign %v, but result was %v", i, tt.src, tt.expected, *tt.dst)
		}
	}
}
