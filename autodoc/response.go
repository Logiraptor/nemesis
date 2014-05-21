package autodoc

import (
	"encoding/json"
	"html/template"
	"net/url"
)

func JSON(x interface{}) Describer {
	return _json{x}
}

type _json struct {
	X interface{}
}

func (j _json) Describe() template.HTML {
	buf, err := json.MarshalIndent(j.X, "    ", "\t")
	if err != nil {
		return template.HTML(err.Error())
	}
	return template.HTML(string(buf))
}

type Values url.Values

func (v Values) Describe() template.HTML {
	return template.HTML(url.Values(v).Encode())
}

type String string

func (s String) Describe() template.HTML {
	return template.HTML(string(s))
}
