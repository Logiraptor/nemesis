package gae

import (
	"errors"
	"reflect"

	"github.com/mjibson/goon"

	"appengine"

	"appengine/datastore"
)

type DB interface {
	Context() appengine.Context
	Count(q *datastore.Query) (int, error)
	Delete(key *datastore.Key) error
	DeleteMulti(keys []*datastore.Key) error
	FlushLocalCache()
	Get(dst interface{}) error
	GetAll(q *datastore.Query, dst interface{}) ([]*datastore.Key, error)
	GetMulti(dst interface{}) error
	Key(src interface{}) *datastore.Key
	KeyError(src interface{}) (*datastore.Key, error)
	Kind(src interface{}) string
	Put(src interface{}) (*datastore.Key, error)
	PutMulti(src interface{}) ([]*datastore.Key, error)
	Run(q *datastore.Query) Iterator
	RunInTransaction(f func(db DB) error, opts *datastore.TransactionOptions) error
}

type Iterator interface {
	Cursor() (datastore.Cursor, error)
	Next(dst interface{}) (*datastore.Key, error)
}

func DBFromContext(ctx appengine.Context) DB {
	return &goonWrapper{
		Goon: goon.FromContext(ctx),
		ctx:  ctx,
	}
}

type goonWrapper struct {
	*goon.Goon
	ctx appengine.Context
}

func (g *goonWrapper) Context() appengine.Context {
	return g.ctx
}

func (g *goonWrapper) Run(q *datastore.Query) Iterator {
	return g.Goon.Run(q)
}

func (g *goonWrapper) RunInTransaction(f func(db DB) error, opts *datastore.TransactionOptions) error {
	return datastore.RunInTransaction(g.ctx, func(c appengine.Context) error {
		db := DBFromContext(c)
		return f(db)
	}, opts)
}

// Put performs validation of the src before calling the standard goon Put method.
func (g *goonWrapper) Put(src interface{}) (*datastore.Key, error) {
	err := Validate(g.ctx, src)
	if err != nil {
		return nil, err
	}
	return g.Goon.Put(src)
}

// PutMulti does validation on each element of a slice before calling the standard goon PutMulti.
// unlike the typical PutMulti, any error during validation causes a rejection of the entire
// source struct.
func (g *goonWrapper) PutMulti(src interface{}) ([]*datastore.Key, error) {
	val := reflect.Indirect(reflect.ValueOf(src))
	if val.Kind() != reflect.Slice {
		return nil, errors.New("src muct be a []T or *[]T")
	}

	elemType := val.Type().Elem()
	if validator, ok := GetValidator(elemType); ok {
		length := val.Len()
		for i := 0; i < length; i++ {
			item := val.Index(i)
			err := validator(g.ctx, item.Interface())
			if err != nil {
				return nil, err
			}
		}
	}

	return g.Goon.PutMulti(src)
}
