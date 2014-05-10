package nemesis

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
)

type testFormStruct struct {
	Email    string
	Password string `forms:",password"`
	Age      int    `forms:"age"`
}

func TestGenerateForm(t *testing.T) {
	Convey("Given a simple user struct", t, func() {
		t := testFormStruct{}
		form, err := GenerateForm(t)
		Convey("The form should generate without error", func() {
			So(err, ShouldBeNil)
		})

		Convey("The form should contain 3 Fields", func() {
			So(len(form.Fields), ShouldEqual, 3)
		})

		Convey("Field type should match struct", func() {
			So(form.Fields[0].Type, ShouldEqual, FormText)
			So(form.Fields[1].Type, ShouldEqual, FormPassword)
			So(form.Fields[2].Type, ShouldEqual, FormNumber)
		})

		Convey("The field names should match the struct", func() {
			So(form.Fields[0].Name, ShouldEqual, "Email")
			So(form.Fields[1].Name, ShouldEqual, "Password")
			So(form.Fields[2].Name, ShouldEqual, "age")
		})
	})
}
