package gae

import (
	"errors"
	"testing"
	. "github.com/smartystreets/goconvey/convey"
	"appengine"
	"appengine/aetest"
	"appengine/datastore"
)

type validatedStruct struct {
	Name   string
	Author *datastore.Key
}

func init() {
	RegisterValidator(&validatedStruct{}, func(c appengine.Context, v interface{}) error {
		val := v.(*validatedStruct)
		if val.Author.Kind() != "Author" {
			return errors.New("Author must be a key to type Author")
		}
		return nil
	})
}

func TestValidateKeyKind(t *testing.T) {
	ctx, err := aetest.NewContext(nil)
	if err != nil {
		t.Fail()
	}
	Convey("Given a struct passing validation rules", t, func() {
		db := DBFromContext(ctx)

		test := &validatedStruct{
			Name:   "Willow",
			Author: datastore.NewKey(ctx, "Author", "", 34235, nil),
		}

		Convey("Err should be nil", func() {
			_, err := db.Put(test)
			So(err, ShouldBeNil)
		})
	})

	Convey("Given a struct failing validation rules", t, func() {
		db := DBFromContext(ctx)

		test := &validatedStruct{
			Name:   "Willow",
			Author: datastore.NewKey(ctx, "Book", "", 3235, nil),
		}

		Convey("Err should be non-nil", func() {
			_, err := db.Put(test)
			So(err, ShouldNotBeNil)
		})
	})

	Convey("Given a set of structs passing validation rules", t, func() {
		db := DBFromContext(ctx)

		test := []*validatedStruct{
			&validatedStruct{
				Name:   "Willow",
				Author: datastore.NewKey(ctx, "Author", "", 323, nil),
			},
			&validatedStruct{
				Name:   "Aspen",
				Author: datastore.NewKey(ctx, "Author", "", 3235, nil),
			},
			&validatedStruct{
				Name:   "Pine",
				Author: datastore.NewKey(ctx, "Author", "", 32335, nil),
			},
		}

		Convey("Err should be nil", func() {
			_, err := db.PutMulti(test)
			So(err, ShouldBeNil)
		})
	})

	Convey("Given a set of structs failing validation rules", t, func() {
		db := DBFromContext(ctx)

		test := []*validatedStruct{
			&validatedStruct{
				Name:   "Willow",
				Author: datastore.NewKey(ctx, "Author", "", 323, nil),
			},
			&validatedStruct{
				Name:   "Aspen",
				Author: datastore.NewKey(ctx, "Nemesis", "", 3235, nil),
			},
			&validatedStruct{
				Name:   "Pine",
				Author: datastore.NewKey(ctx, "Author", "", 32335, nil),
			},
		}

		Convey("Err should be nil", func() {
			_, err := db.PutMulti(test)
			So(err, ShouldNotBeNil)
		})
	})
}
