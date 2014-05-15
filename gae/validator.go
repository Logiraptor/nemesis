package gae

import (
	"appengine"

	"reflect"
)

type Validator func(c appengine.Context, v interface{}) error

var validators map[reflect.Type]Validator

func RegisterValidator(typ interface{}, v Validator) {
	if validators == nil {
		validators = make(map[reflect.Type]Validator)
	}

	validators[reflect.TypeOf(typ)] = v
}

func Validate(ctx appengine.Context, src interface{}) error {
	t := reflect.TypeOf(src)
	if v, ok := validators[t]; ok {
		return v(ctx, src)
	}
	return nil
}

func GetValidator(src reflect.Type) (Validator, bool) {
	v, ok := validators[src]
	return v, ok
}
