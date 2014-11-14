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
	if len(tokens) == 0 { //*|| len(tokens[0]) == 0 {
		return nil, fmt.Errorf("Invalid JSON pointer: %q", pointer)
	}
	return tokens, nil
}

func Has(obj interface{}, pointer string) (rv bool) {
	defer func() {
		if e := recover(); e != nil {
			rv = false
		}
	}()
	tokens, err := parse(pointer)
	if err != nil {
		return false
	}

	i := 0
	v := reflect.ValueOf(obj)
	if len(tokens) > 0 && tokens[0] != "" {
		for i < len(tokens) {
			for v.Kind() == reflect.Interface {
				v = v.Elem()
			}
			token := tokens[i]
			if n, err := strconv.Atoi(token); err == nil {
				v = v.Index(n)
			} else {
				v = v.MapIndex(reflect.ValueOf(token))
			}
			i++
		}
		return v.IsValid()
	}
	return false
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
	if len(tokens) > 0 && tokens[0] != "" {
		for i < len(tokens) {
			for v.Kind() == reflect.Interface {
				v = v.Elem()
			}
			token := tokens[i]
			if n, err := strconv.Atoi(token); err == nil {
				v = v.Index(n)
			} else {
				v = v.MapIndex(reflect.ValueOf(token))
			}
			i++
		}
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
	if len(tokens) > 0 && tokens[0] != "" {
		for i < len(tokens) {
			for v.Kind() == reflect.Interface {
				v = v.Elem()
			}
			p = v
			token = tokens[i]
			if n, err := strconv.Atoi(token); err == nil {
				v = v.Index(n)
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
	return fmt.Errorf("pointer should have element")
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
	if len(tokens) > 0 && tokens[0] != "" {
		for i < len(tokens) {
			for v.Kind() == reflect.Interface {
				v = v.Elem()
			}
			pp = p
			p = v
			ptoken = token
			token = tokens[i]
			if n, err := strconv.Atoi(token); err == nil {
				v = v.Index(n)
			} else {
				v = v.MapIndex(reflect.ValueOf(token))
			}
			i++
		}
	} else {
		return nil, fmt.Errorf("pointer should have element")
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
		nv = reflect.Zero(p.Type())
		n, _ := strconv.Atoi(token)
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
		p.Set(reflect.ValueOf(nv))
	}
	return obj, nil
}
