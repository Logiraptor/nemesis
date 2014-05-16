package nemesis

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/codegangsta/martini"

	"io"
)

type JSONEncoder struct {
	wr io.Writer
}

type Error struct {
	Error string
}

// Encode attempts to write v as JSON. if v is an error, the json is formatted as
// {"Error":%message%}, otherwise the default json encoding is written.
func (e *JSONEncoder) Encode(v interface{}) error {
	if err, ok := v.(error); ok {
		return json.NewEncoder(e.wr).Encode(map[string]string{
			"Error": err.Error(),
		})
	}
	return json.NewEncoder(e.wr).Encode(v)
}

// EncodeResponse takes a function of the form
// func(...) (int, interface{}) and encodes the return value
// to the ResponseWriter as JSON
func EncodeResponse(f interface{}) interface{} {
	t := reflect.TypeOf(f)
	if t.NumOut() != 2 {
		panic("func passed to EncodeResponse should return (int, interface{})")
	}

	first := t.Out(0)
	if first.Kind() != reflect.Int {
		panic(fmt.Sprintf("first return value is %s, expected int", first.String()))
	}

	second := t.Out(1)
	if second.Kind() != reflect.Interface {
		panic(fmt.Sprintf("first return value is %s, expected interface{}", second.String()))
	}

	return func(c martini.Context, rw http.ResponseWriter) {
		enc := JSONEncoder{rw}
		result, err := c.Invoke(f)
		if err != nil {
			enc.Encode(err)
			return
		}

		rw.Header().Add("Content-Type", "application/json")
		rw.WriteHeader(int(result[0].Int()))
		v := result[1].Interface()
		if v == nil {
			return
		}
		err = enc.Encode(v)
		if err != nil {
			enc.Encode(err)
			return
		}
	}
}

// EncodeResponse takes a function of the form
// func(...) (interface{}, error) and encodes the return value
// to the ResponseWriter as JSON. If error != nil, 500 is returned,
// otherwise 200.
func EncodeResponseError(f interface{}) interface{} {
	t := reflect.TypeOf(f)
	if t.NumOut() != 2 {
		panic("func passed to EncodeResponse should return (interface{}, error)")
	}

	first := t.Out(0)
	if first.Kind() != reflect.Interface {
		panic(fmt.Sprintf("first return value is %s, expected interface{}", first.String()))
	}

	second := t.Out(1)
	if second.Kind() != reflect.Interface {
		panic(fmt.Sprintf("first return value is %s, expected error", second.String()))
	}

	return func(c martini.Context, rw http.ResponseWriter) {
		rw.Header().Add("Content-Type", "application/json")
		enc := JSONEncoder{rw}
		result, err := c.Invoke(f)
		if err != nil {
			rw.WriteHeader(500)
			enc.Encode(err)
			return
		}

		v := result[0].Interface()
		e := result[1].Interface()
		if err, ok := e.(error); ok && err != nil {
			rw.WriteHeader(500)
			enc.Encode(err)
			return
		}
		if v == nil {
			return
		}
		rw.WriteHeader(200)
		err = enc.Encode(v)
		if err != nil {
			enc.Encode(err)
			return
		}
	}
}
