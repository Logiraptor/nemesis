package nemesis

import (
	"net/url"
	"testing"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFakePost(t *testing.T) {
	Convey("Given a parameter map", t, func() {
		val := url.Values{
			"username": {"test@mail.com"},
			"password": {"hunter2"},
		}

		Convey("The request should match the map", func() {
			req, err := ForgeFormPost("/", val)

			So(err, ShouldBeNil)
			So(req.FormValue("username"), ShouldEqual, "test@mail.com")
			So(req.FormValue("password"), ShouldEqual, "hunter2")
		})
	})

	Convey("Given a bad request", t, func() {
		val := url.Values{
			"username": {"test@mail.com"},
			"password": {"hunter2"},
		}

		Convey("The request should err", func() {
			_, err := ForgeFormPost("$$", val)
			So(err, ShouldNotBeNil)
		})
	})
}
