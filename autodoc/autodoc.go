package autodoc

import (
	"encoding/json"
	"html/template"
	"strings"

	"net/http"
)

func init() {
	var err error
	docTempl, err = template.New("doc").Funcs(template.FuncMap{
		"id": func(x string) string {
			return strings.Replace(x, " ", "_", -1)
		},
		"json": func(x ...interface{}) template.HTML {
			buf, err := json.MarshalIndent(x[0], "    ", "\t")
			if err != nil {
				return template.HTML(err.Error())
			}
			return template.HTML(string(buf))
		},
	}).Parse(base_html)
	if err != nil {
		panic(err)
	}
}

var docTempl *template.Template

type APIDoc struct {
	URL, Method       string
	Name, Description string
	Request           Describer
	Response          Describer
}

type APIDocList struct {
	Title, Root string
	Theme       string
	APIs        []APIDoc
}

func NewAPIDocList(name, root, theme string) *APIDocList {
	return &APIDocList{
		Title: name,
		Root:  root,
		Theme: theme,
	}
}

func (a *APIDocList) AddMethods(doc ...APIDoc) {
	for i := range doc {
		doc[i].URL = a.Root + doc[i].URL
	}
	a.APIs = append(a.APIs, doc...)
}

func (a *APIDocList) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	err := docTempl.Execute(rw, a)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}
