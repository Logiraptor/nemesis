package nemesis

import "reflect"

type FormType string

const (
	FormText     FormType = "text"
	FormNumber   FormType = "number"
	FormPassword FormType = "password"
)

type Field struct {
	Type    FormType
	Name    string
	Default string
	Label   string
}

type AutoForm struct {
	Fields []Field
	Action string
	Method string
}

func GenerateForm(x interface{}) (*AutoForm, error) {
	t := reflect.TypeOf(x)
	numFields := t.NumField()
	var form = new(AutoForm)
	form.Fields = make([]Field, numFields)

	for i := range form.Fields {
		structField := t.Field(i)
		form.Fields[i].Name = structField.Name
		typ := structField.Type.Kind()
		switch typ {
		case reflect.String:
			form.Fields[i].Type = FormText
		case reflect.Int, reflect.Float32, reflect.Float64:
			form.Fields[i].Type = FormNumber
		}

		tag := structField.Tag.Get("forms")
		mods, _ := ParseStructTags(tag, []string{"name", "type"})
		if name, ok := mods["name"]; ok {
			form.Fields[i].Name = name
		}
		if typ, ok := mods["type"]; ok {
			form.Fields[i].Type = FormType(typ)
		}
	}

	return form, nil
}
