package autodoc

import (
	"html/template"
)

type Describer interface {
	Describe() template.HTML
}
