package jsonpointer

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// parse splits a RFC 6901 JSON pointer into its reference tokens.
// The empty string is a valid pointer that references the whole document and
// yields zero tokens. Any other pointer must begin with '/'.
func parse(pointer string) ([]string, error) {
	if pointer == "" {
		return nil, nil
	}
	if !strings.HasPrefix(pointer, "/") {
		return nil, fmt.Errorf("invalid JSON pointer: %q", pointer)
	}
	tokens := strings.Split(pointer[1:], "/")
	for i, token := range tokens {
		// '~' must always be followed by '0' or '1' (RFC 6901 sec.3).
		for j := 0; j < len(token); j++ {
			if token[j] == '~' {
				if j+1 >= len(token) || (token[j+1] != '0' && token[j+1] != '1') {
					return nil, fmt.Errorf("invalid escape in JSON pointer: %q", pointer)
				}
				j++
			}
		}
		// Unescape '~1' to '/' first, then '~0' to '~' (order matters).
		tokens[i] = strings.Replace(
			strings.Replace(token, "~1", "/", -1), "~0", "~", -1)
	}
	return tokens, nil
}

// arrayIndex reports whether token is a valid RFC 6901 array index, i.e. "0"
// or a non-zero-prefixed sequence of digits. "-", "+1", "01" and the like are
// rejected.
func arrayIndex(token string) (int, bool) {
	if token == "0" {
		return 0, true
	}
	if len(token) == 0 || token[0] < '1' || token[0] > '9' {
		return 0, false
	}
	for i := 1; i < len(token); i++ {
		if token[i] < '0' || token[i] > '9' {
			return 0, false
		}
	}
	n, err := strconv.Atoi(token)
	if err != nil {
		return 0, false
	}
	return n, true
}

// eval walks tokens from v and returns the referenced value. It panics (via
// reflect) when a token references a nonexistent member, which callers recover.
func eval(v reflect.Value, tokens []string) reflect.Value {
	for _, token := range tokens {
		for v.Kind() == reflect.Interface {
			v = v.Elem()
		}
		if isIndexed(v) {
			n, ok := arrayIndex(token)
			if !ok {
				panic(fmt.Sprintf("invalid array index: %q", token))
			}
			v = v.Index(n)
		} else {
			v = v.MapIndex(reflect.ValueOf(token))
		}
	}
	return v
}

// Has return whether the obj has pointer.
func Has(obj any, pointer string) (rv bool) {
	defer func() {
		if e := recover(); e != nil {
			rv = false
		}
	}()
	tokens, err := parse(pointer)
	if err != nil {
		return false
	}
	return eval(reflect.ValueOf(obj), tokens).IsValid()
}

// Get return a value which is pointed with pointer on obj.
func Get(obj any, pointer string) (rv any, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("invalid JSON pointer: %q: %v", pointer, e)
		}
	}()
	tokens, err := parse(pointer)
	if err != nil {
		return nil, err
	}
	v := eval(reflect.ValueOf(obj), tokens)
	if !v.IsValid() {
		return nil, fmt.Errorf("invalid JSON pointer: %q", pointer)
	}
	return v.Interface(), nil
}

// Set set a value which is pointed with pointer on obj.
func Set(obj any, pointer string, value any) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("invalid JSON pointer: %q: %v", pointer, e)
		}
	}()
	tokens, err := parse(pointer)
	if err != nil {
		return err
	}
	if len(tokens) == 0 {
		return fmt.Errorf("pointer should have element")
	}

	v := reflect.ValueOf(obj)
	var pp reflect.Value
	var ptoken string
	for _, token := range tokens[:len(tokens)-1] {
		for v.Kind() == reflect.Interface {
			v = v.Elem()
		}
		pp, ptoken = v, token
		if n, ok := arrayIndex(token); ok && isIndexed(v) {
			v = v.Index(n)
		} else {
			v = v.MapIndex(reflect.ValueOf(token))
		}
	}
	for v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	p := v
	token := tokens[len(tokens)-1]

	if p.Kind() == reflect.Map {
		rv := reflect.ValueOf(value)
		if value == nil {
			rv = reflect.Zero(p.Type().Elem())
		}
		p.SetMapIndex(reflect.ValueOf(token), rv)
		return nil
	}
	if isIndexed(p) {
		// "-" references the (nonexistent) element after the last one and
		// means "append" in a Set context (RFC 6901 sec.4).
		if token == "-" {
			rv := reflect.ValueOf(value)
			if value == nil {
				rv = reflect.Zero(p.Type().Elem())
			}
			return writeBack(pp, ptoken, reflect.Append(p, rv))
		}
		n, ok := arrayIndex(token)
		if !ok {
			return fmt.Errorf("invalid array index: %q", token)
		}
		el := p.Index(n)
		rv := reflect.ValueOf(value)
		if value == nil {
			rv = reflect.Zero(el.Type())
		}
		el.Set(rv)
		return nil
	}
	return fmt.Errorf("invalid JSON pointer: %q", pointer)
}

// writeBack stores nv into the parent container pp at ptoken. It is used when a
// container value is replaced wholesale (e.g. growing a slice on append).
func writeBack(pp reflect.Value, ptoken string, nv reflect.Value) error {
	if !pp.IsValid() {
		return fmt.Errorf("cannot grow the root array")
	}
	for pp.Kind() == reflect.Interface {
		pp = pp.Elem()
	}
	if pp.Kind() == reflect.Map {
		pp.SetMapIndex(reflect.ValueOf(ptoken), nv)
		return nil
	}
	n, ok := arrayIndex(ptoken)
	if !ok {
		return fmt.Errorf("invalid array index: %q", ptoken)
	}
	pp.Index(n).Set(nv)
	return nil
}

// Remove remove a value which is pointed with pointer on obj.
func Remove(obj any, pointer string) (rv any, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("invalid JSON pointer: %q: %v", pointer, e)
		}
	}()
	tokens, err := parse(pointer)
	if err != nil {
		return nil, err
	}
	if len(tokens) == 0 {
		return nil, fmt.Errorf("pointer should have element")
	}

	v := reflect.ValueOf(obj)
	var p, pp reflect.Value
	var token, ptoken string
	for i := 0; i < len(tokens); i++ {
		for v.Kind() == reflect.Interface {
			v = v.Elem()
		}
		pp, ptoken = p, token
		p, token = v, tokens[i]
		if n, ok := arrayIndex(token); ok && isIndexed(v) {
			v = v.Index(n)
		} else {
			v = v.MapIndex(reflect.ValueOf(token))
		}
	}

	var nv reflect.Value
	if p.Kind() == reflect.Map {
		nv = reflect.MakeMap(p.Type())
		for _, mk := range p.MapKeys() {
			if mk.String() != token {
				nv.SetMapIndex(mk, p.MapIndex(mk))
			}
		}
	} else {
		nv = reflect.MakeSlice(p.Type(), 0, p.Len())
		n, _ := arrayIndex(token)
		for m := 0; m < p.Len(); m++ {
			if n != m {
				nv = reflect.Append(nv, p.Index(m))
			}
		}
	}

	if !pp.IsValid() {
		obj = nv.Interface()
	} else if pp.Kind() == reflect.Map {
		pp.SetMapIndex(reflect.ValueOf(ptoken), nv)
	} else {
		n, _ := arrayIndex(ptoken)
		pp.Index(n).Set(nv)
	}
	return obj, nil
}

func isIndexed(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array:
		return true
	case reflect.Slice:
		return true
	}
	return false
}
