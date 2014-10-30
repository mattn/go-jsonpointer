# go-jsonpointer

Go implementation of JSON Pointer (RFC6901)

## Usage

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

## Installation

```
$ go get github.com/mattn/go-jsonpath
```

## License

MIT

## Author

Yasuhiro Matsumoto (a.k.a mattn)
