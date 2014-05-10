package nemesis

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
)

func TestParseStructTags(t *testing.T) {
	Convey("Given a struct tag and schema", t, func() {
		tag := "foo,,nope"
		schema := []string{
			"name",
			"length",
			"index",
			"other",
		}

		val, err := ParseStructTags(tag, schema)

		So(err, ShouldBeNil)
		So(val, ShouldResemble, map[string]string{
			"name":  "foo",
			"index": "nope",
		})
	})

	Convey("Given a struct tag and schema", t, func() {
		tag := ""
		schema := []string{
			"name",
			"length",
			"index",
			"other",
		}

		val, err := ParseStructTags(tag, schema)

		So(err, ShouldBeNil)
		So(val, ShouldResemble, map[string]string{})
	})
}
