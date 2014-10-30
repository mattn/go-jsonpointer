package jsonpointer

import (
	"encoding/json"
	"reflect"
	"testing"
)

var testGetCases = []struct {
	json   string
	path   string
	expect interface{}
	err    string
}{
	{`{"foo":[1,3,true]}`, `/foo/2`, true, ``},
	{`{"foo":2}`, `/foo`, 2.0, ``},
	{`{"foo":[]}`, `/foo`, []interface{}{}, ``},
	{`{"foo":"yes"}`, `/foo`, "yes", ``},
	{`{"foo":3.14}`, `/`, "", `Invalid JSON pointer: "/"`},
}

func TestGet(t *testing.T) {
	for _, testcase := range testGetCases {
		var obj interface{}
		err := json.Unmarshal([]byte(testcase.json), &obj)
		if err != nil {
			t.Fatal(err)
		}

		value, err := Get(obj, testcase.path)
		if err != nil {
			if err.Error() != testcase.err {
				t.Fatal(err)
			}
		} else if !reflect.DeepEqual(value, testcase.expect) {
			t.Fatalf("Expected %v, but %v:", testcase.expect, value)
		}
	}
}

var testSetCases = []struct {
	json   string
	path   string
	value  interface{}
	expect string
	err    string
}{
	{`{"foo":[1,3,true]}`, `/foo/2`, "false", `{"foo":[1,3,"false"]}`, ``},
	{`{"foo":2}`, `/foo`, "true", `{"foo":"true"}`, ``},
	{`{"foo":2}`, `/foo`, true, `{"foo":true}`, ``},
	{`{"foo":2}`, `/foo`, "2", `{"foo":"2"}`, ``},
	{`{"foo":3.14}`, `/foo`, 1.5, `{"foo":1.5}`, ``},
	{`{"foo":3.14}`, `/`, 1.5, `{}`, `Invalid JSON pointer: "/"`},
}

func TestSet(t *testing.T) {
	for _, testcase := range testSetCases {
		var obj, expect interface{}
		err := json.Unmarshal([]byte(testcase.json), &obj)
		if err != nil {
			t.Fatal(err)
		}
		err = json.Unmarshal([]byte(testcase.expect), &expect)
		if err != nil {
			t.Fatal(err)
		}

		err = Set(obj, testcase.path, testcase.value)
		if err != nil {
			if err.Error() != testcase.err {
				t.Fatal(err)
			}
		} else if !reflect.DeepEqual(obj, expect) {
			t.Fatalf("Expected %v, but %v:", expect, obj)
		}
	}
}

var testRemoveCases = []struct {
	json   string
	path   string
	expect string
	err    string
}{
	{`{"foo":2,"bar":3}`, `/bar`, `{"foo":2}`, ``},
	{`{"foo":[1,3,true]}`, `/foo/1`, `{"foo":[1,true]}`, ``},
	{`{"foo":[]}`, `/foo`, `{}`, ``},
	{`{"foo":3.14}`, `/`, `{}`, `Invalid JSON pointer: "/"`},
}

func TestRemove(t *testing.T) {
	for _, testcase := range testRemoveCases {
		var obj, expect interface{}
		err := json.Unmarshal([]byte(testcase.json), &obj)
		if err != nil {
			t.Fatal(err)
		}
		err = json.Unmarshal([]byte(testcase.expect), &expect)
		if err != nil {
			t.Fatal(err)
		}

		v, err := Remove(obj, testcase.path)
		if err != nil {
			if err.Error() != testcase.err {
				t.Fatal(err)
			}
		} else if !reflect.DeepEqual(v, expect) {
			t.Fatalf("Expected %v, but %v:", expect, v)
		}
	}
}