# go-jsonpointer

Go implementation of JSON Pointer (RFC6901)

## Usage

`jsonpointer.Get(obj, pointer)`
```go
json := `
{
	"foo": [1,true,2]
}
`
var obj interface{}
json.Unmarshal([]byte(json), &obj)
jsonpointer.Get(obj, "/foo/1") // Should be true
```

`jsonpointer.Set(obj, pointer, newvalue)`
```go
json := `
{
	"foo": [1,true,2]
}
`
var obj interface{}
json.Unmarshal([]byte(json), &obj)
jsonpointer.Set(obj, "/foo/1", false)
// obj should be {"foo":[1,false,2]}
```

`jsonpointer.Remove(obj, pointer)`
```go
json := `
{
	"foo": [1,true,2]
}
`
var obj interface{}
json.Unmarshal([]byte(json), &obj)
jsonpointer.Remove(obj, "/foo/1")
// obj should be {"foo":[1,2]}
```

## Installation

```
$ go get github.com/mattn/go-jsonpointer
```

## License

MIT

## Author

Yasuhiro Matsumoto (a.k.a mattn)
