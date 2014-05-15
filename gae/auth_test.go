package gae

import (
	"net/url"
	"testing"

	"github.com/mjibson/goon"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/Logiraptor/nemesis"

	"appengine/aetest"
)

func TestAddUser(t *testing.T) {
	Convey("Given a registration request", t, func() {
		c, err := aetest.NewContext(nil)
		if err != nil {
			t.Errorf("%s", err.Error())
		}
		defer c.Close()

		req, err := nemesis.ForgeFormPost("/", url.Values{
			"username": {"test@gmail.com"},
			"password": {"password123"},
		})
		var u User = new(BaseUser)
		u, err = CreateUser(c, req, u)
		So(err, ShouldBeNil)
		So(u.Username(), ShouldEqual, "test@gmail.com")
	})
}

type userMeta struct {
	BaseUser
	Id            int64 `goon:"id"`
	FavoriteCream string
}

func TestAddMetaData(t *testing.T) {
	Convey("Given a registration request", t, func() {
		c, err := aetest.NewContext(nil)
		if err != nil {
			t.Errorf("%s", err.Error())
		}
		defer c.Close()

		req, err := nemesis.ForgeFormPost("/", url.Values{
			"username": {"test@gmail.com"},
			"password": {"password123"},
		})
		var u = new(userMeta)
		u.FavoriteCream = "French Vanilla"
		_, err = CreateUser(c, req, u)
		So(err, ShouldBeNil)
		So(u.Id, ShouldNotEqual, 0)
		So(u.Username(), ShouldEqual, "test@gmail.com")
		So(u.FavoriteCream, ShouldEqual, "French Vanilla")

		recall := new(userMeta)
		recall.Id = u.Id

		g := goon.FromContext(c)
		err = g.Get(recall)
		So(err, ShouldBeNil)
		So(u.Id, ShouldNotEqual, 0)
		So(u.Username(), ShouldEqual, "test@gmail.com")
		So(u.FavoriteCream, ShouldEqual, "French Vanilla")
	})
}
