package jsonpointer

import (
	"encoding/json"
	"reflect"
	"testing"
)

var testHasCases = []struct {
	json    string
	pointer string
	expect  bool
}{
	{`{"foo":[1,3,true]}`, `/foo/2`, true},
	{`{"foo":[1,3,true]}`, `/foo/3`, false},
	{`{"foo":2}`, `/foo`, true},
	{`{"foo":[]}`, `/fooo`, false},
	{`{"foo":3.14}`, ``, true},
	{`{"hoge":"fuga","foo":{"fuga":"foo1","hoge":"foo2"}}`, `/foo/fuga`, true},
	{`{"foo~bar/baz":[1,3,true]}`, `/foo~0bar~1baz/1`, true},
	{`{"0": [9, 8, 7]}`, `/0/1`, true},
	{`{"0": {"foo": 8}}`, `/0/foo`, true},
	{`{"0": {" ": "foo"}}`, `/0/ `, true},
}

func TestHas(t *testing.T) {
	for _, testcase := range testHasCases {
		var obj any
		err := json.Unmarshal([]byte(testcase.json), &obj)
		if err != nil {
			t.Fatal(err)
		}

		value := Has(obj, testcase.pointer)
		if value != testcase.expect {
			t.Fatalf("expected %v, but %v:", testcase.expect, value)
		}
	}
}

var testGetCases = []struct {
	json    string
	pointer string
	expect  any
	err     string
}{
	{`{"foo":[1,3,true]}`, `/foo/2`, true, ``},
	{`{"foo":2}`, `/foo`, 2.0, ``},
	{`{"foo":[]}`, `/foo`, []any{}, ``},
	{`{"foo":"yes"}`, `/foo`, "yes", ``},
	{`{"foo":3.14}`, ``, map[string]any{"foo": 3.14}, ``},
	{`{"foo":3.14}`, `/`, nil, `invalid JSON pointer: "/"`},
	{`{"":7}`, `/`, 7.0, ``},
	{`{"hoge":"fuga","foo":{"fuga":"foo1","hoge":"foo2"}}`, `/foo/fuga`, "foo1", ``},
	{`{"foo~bar/baz":[1,3,true]}`, `/foo~0bar~1baz/1`, 3.0, ``},
	{`{"0": [9, 8, 7]}`, `/0/1`, 8.0, ``},
	{`{"0": {" ": "foo"}}`, `/0/ `, "foo", ``},
}

func TestGet(t *testing.T) {
	for _, testcase := range testGetCases {
		var obj any
		err := json.Unmarshal([]byte(testcase.json), &obj)
		if err != nil {
			t.Fatal(err)
		}

		value, err := Get(obj, testcase.pointer)
		if err != nil {
			if err.Error() != testcase.err {
				t.Fatal(testcase.json, err)
			}
		} else if !reflect.DeepEqual(value, testcase.expect) {
			t.Fatalf("Expected %v, but %v:", testcase.expect, value)
		}
	}
}

var testSetCases = []struct {
	json    string
	pointer string
	value   any
	expect  string
	err     string
}{
	{`{"foo":[1,3,true]}`, `/foo/2`, "false", `{"foo":[1,3,"false"]}`, ``},
	{`{"foo":2}`, `/foo`, "true", `{"foo":"true"}`, ``},
	{`{"foo":2}`, `/foo`, true, `{"foo":true}`, ``},
	{`{"foo":2}`, `/foo`, "2", `{"foo":"2"}`, ``},
	{`{"foo":3.14}`, `/foo`, 1.5, `{"foo":1.5}`, ``},
	{`{"foo":3.14}`, `/`, 1.5, `{"foo":3.14,"":1.5}`, ``},
	{`{"hoge":"fuga","foo":{"fuga":"foo1","hoge":"foo2"}}`, `/foo/fuga`, 3.0, `{"hoge":"fuga","foo":{"fuga":3,"hoge":"foo2"}}`, ``},
	{`{"foo~bar/baz":[1,3,true]}`, `/foo~0bar~1baz/1`, 4.0, `{"foo~bar/baz":[1,4,true]}`, ``},
	{`{"0": [9, 8, 7]}`, `/0/1`, 20.0, `{"0": [9, 20, 7]}`, ``},
	{`{"0": {" ": "foo"}}`, `/0/ `, "bar", `{"0": {" ": "bar"}}`, ``},
	{`[[1,2],[3,4]]`, `/1/0`, 9.0, `[[1,2],[9,4]]`, ``},
	{`{"foo":1}`, `/foo`, nil, `{"foo":null}`, ``},
	{`{"foo":[1,2,3]}`, `/foo/1`, nil, `{"foo":[1,null,3]}`, ``},
	{`{"foo":1}`, `/bar`, 5.0, `{"foo":1,"bar":5}`, ``},
	{`{"a":[1,2]}`, `/a/-`, 3.0, `{"a":[1,2,3]}`, ``},
	{`{"a":[[1,2],[3,4]]}`, `/a/0/-`, 9.0, `{"a":[[1,2,9],[3,4]]}`, ``},
}

func TestSet(t *testing.T) {
	for _, testcase := range testSetCases {
		var obj, expect any
		err := json.Unmarshal([]byte(testcase.json), &obj)
		if err != nil {
			t.Fatal(err)
		}
		err = json.Unmarshal([]byte(testcase.expect), &expect)
		if err != nil {
			t.Fatal(err)
		}

		err = Set(obj, testcase.pointer, testcase.value)
		if err != nil {
			if err.Error() != testcase.err {
				t.Fatal(err)
			}
		} else if !reflect.DeepEqual(obj, expect) {
			t.Fatalf("expected %v, but %v:", expect, obj)
		}
	}
}

var testRemoveCases = []struct {
	json    string
	pointer string
	expect  string
	err     string
}{
	{`{"foo":2,"bar":3}`, `/bar`, `{"foo":2}`, ``},
	{`{"foo":[1,3,true]}`, `/foo/1`, `{"foo":[1,true]}`, ``},
	{`{"foo":[]}`, `/foo`, `{}`, ``},
	{`{"foo":3.14}`, `/`, `{"foo":3.14}`, ``},
	{`{"hoge":"fuga","foo":{"fuga":"foo1","hoge":"foo2"}}`, `/foo/fuga`, `{"hoge":"fuga","foo":{"hoge":"foo2"}}`, ``},
	{`{"foo~bar/baz":[1,3,true]}`, `/foo~0bar~1baz/1`, `{"foo~bar/baz":[1,true]}`, ``},
	{`{"0": [9, 8, 7]}`, `/0/1`, `{"0": [9, 7]}`, ``},
	{`{"0": {" ": "foo"}}`, `/0/ `, `{"0": {}}`, ``},
	{`[[1,2,3],[4,5,6]]`, `/0/1`, `[[1,3],[4,5,6]]`, ``},
	{`{"foo":[1]}`, `/foo/0`, `{"foo":[]}`, ``},
}

func TestRemove(t *testing.T) {
	for _, testcase := range testRemoveCases {
		var obj, expect any
		err := json.Unmarshal([]byte(testcase.json), &obj)
		if err != nil {
			t.Fatal(err)
		}
		err = json.Unmarshal([]byte(testcase.expect), &expect)
		if err != nil {
			t.Fatal(err)
		}

		v, err := Remove(obj, testcase.pointer)
		if err != nil {
			if err.Error() != testcase.err {
				t.Fatal(err)
			}
		} else if !reflect.DeepEqual(v, expect) {
			t.Fatalf("expected %v, but %v:", expect, v)
		}
	}
}

func obj(t *testing.T, s string) any {
	var o any
	if err := json.Unmarshal([]byte(s), &o); err != nil {
		t.Fatal(err)
	}
	return o
}

// TestRFC6901 checks compliance details whose exact error messages are not
// worth pinning in the table tests.
func TestRFC6901(t *testing.T) {
	// "" references the whole document.
	if v, err := Get(obj(t, `{"foo":1}`), ``); err != nil || !reflect.DeepEqual(v, obj(t, `{"foo":1}`)) {
		t.Errorf(`Get "" => %v, %v`, v, err)
	}
	// "/" references the member with the empty key.
	if v, err := Get(obj(t, `{"":5}`), `/`); err != nil || v != 5.0 {
		t.Errorf(`Get "/" => %v, %v`, v, err)
	}
	// A reference token that fails to resolve is an error.
	mustErr := func(p string, err error) {
		t.Helper()
		if err == nil {
			t.Errorf("Get %q: expected error, got nil", p)
		}
	}
	// Array indexes must not have leading zeros or signs, and "-" is not
	// readable.
	for _, p := range []string{`/a/01`, `/a/+1`, `/a/-`, `/a/-1`} {
		_, err := Get(obj(t, `{"a":[10,20,30]}`), p)
		mustErr(p, err)
	}
	// Leading whitespace and bad escapes are invalid pointers.
	for _, p := range []string{` /foo`, `foo`, `/~`, `/~2`} {
		_, err := Get(obj(t, `{"foo":1}`), p)
		mustErr(p, err)
	}
}
