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
	{`{"foo":3.14}`, ``, false},
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
	{`{"foo":3.14}`, ``, "", `invalid JSON pointer: ""`},
	{`{"foo":3.14}`, `/`, map[string]any{"foo": 3.14}, ``},
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
	{`{"foo":3.14}`, `/`, 1.5, `{}`, `pointer should have element`},
	{`{"hoge":"fuga","foo":{"fuga":"foo1","hoge":"foo2"}}`, `/foo/fuga`, 3.0, `{"hoge":"fuga","foo":{"fuga":3,"hoge":"foo2"}}`, ``},
	{`{"foo~bar/baz":[1,3,true]}`, `/foo~0bar~1baz/1`, 4.0, `{"foo~bar/baz":[1,4,true]}`, ``},
	{`{"0": [9, 8, 7]}`, `/0/1`, 20.0, `{"0": [9, 20, 7]}`, ``},
	{`{"0": {" ": "foo"}}`, `/0/ `, "bar", `{"0": {" ": "bar"}}`, ``},
	{`[[1,2],[3,4]]`, `/1/0`, 9.0, `[[1,2],[9,4]]`, ``},
	{`{"foo":1}`, `/foo`, nil, `{"foo":null}`, ``},
	{`{"foo":[1,2,3]}`, `/foo/1`, nil, `{"foo":[1,null,3]}`, ``},
	{`{"foo":1}`, `/bar`, 5.0, `{"foo":1,"bar":5}`, ``},
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
	{`{"foo":3.14}`, `/`, `{}`, `pointer should have element`},
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
