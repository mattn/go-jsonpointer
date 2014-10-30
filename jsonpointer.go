package jsonpointer

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func parse(pointer string) ([]string, error) {
	pointer = strings.TrimSpace(pointer)
	if !strings.HasPrefix(pointer, "/") {
		return nil, fmt.Errorf("Invalid JSON pointer: %q", pointer)
	}
	tokens := strings.Split(pointer[1:], "/")
	if len(tokens) == 0 || len(tokens[0]) == 0 {
		return nil, fmt.Errorf("Invalid JSON pointer: %q", pointer)
	}
	return tokens, nil
}

func Get(obj interface{}, pointer string) (rv interface{}, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("Invalid JSON pointer: %q: %v", pointer, e)
		}
	}()
	tokens, err := parse(pointer)
	if err != nil {
		return nil, err
	}

	i := 0
	v := reflect.ValueOf(obj)
	for i < len(tokens) {
		token := tokens[i]
		if n, err := strconv.Atoi(token); err == nil {
			v = v.Elem().Index(n)
		} else {
			v = v.MapIndex(reflect.ValueOf(token))
		}
		i++
	}
	return v.Interface(), nil
}

func Set(obj interface{}, pointer string, value interface{}) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("Invalid JSON pointer: %q: %v", pointer, e)
		}
	}()
	tokens, err := parse(pointer)
	if err != nil {
		return err
	}

	i := 0
	v := reflect.ValueOf(obj)
	var p reflect.Value
	var token string
	for i < len(tokens) {
		p = v
		token = tokens[i]
		if n, err := strconv.Atoi(token); err == nil {
			v = v.Elem().Index(n)
		} else {
			v = v.MapIndex(reflect.ValueOf(token))
		}
		i++
	}

	if p.Kind() == reflect.Map {
		p.SetMapIndex(reflect.ValueOf(token), reflect.ValueOf(value))
	} else {
		v.Set(reflect.ValueOf(value))
	}
	return nil
}

func Remove(obj interface{}, pointer string) (rv interface{}, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("Invalid JSON pointer: %q: %v", pointer, e)
		}
	}()
	tokens, err := parse(pointer)
	if err != nil {
		return nil, err
	}

	i := 0
	v := reflect.ValueOf(obj)
	var p, pp reflect.Value
	var token, ptoken string
	for i < len(tokens) {
		pp = p
		p = v
		ptoken = token
		token = tokens[i]
		if n, err := strconv.Atoi(token); err == nil {
			v = v.Elem().Index(n)
		} else {
			v = v.MapIndex(reflect.ValueOf(token))
		}
		i++
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
		nv = reflect.Zero(p.Elem().Type())
		n, _ := strconv.Atoi(token)
		for m := 0; m < p.Elem().Len(); m++ {
			if n != m {
				nv = reflect.Append(nv, p.Elem().Index(m))
			}
		}
	}

	if !pp.IsValid() {
		obj = nv.Interface()
	} else if pp.Kind() == reflect.Map {
		pp.SetMapIndex(reflect.ValueOf(ptoken), nv)
	} else {
		p.Set(reflect.ValueOf(nv))
	}
	return obj, nil
}