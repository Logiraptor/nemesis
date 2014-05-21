package main

import (
	"net/http"

	"github.com/Logiraptor/nemesis/autodoc"
)

func main() {
	doc := autodoc.NewAPIDocList("Example AutoDoc", "/", "cerulean")
	doc.AddMethods(
		autodoc.APIDoc{
			Method:      "GET",
			Name:        "Endpoint A",
			Description: "Get all the A through json",
			Request: autodoc.JSON(map[string]interface{}{
				"Test": 2,
			}),
			Response: autodoc.JSON(map[string]interface{}{
				"Test": 2,
			}),
		},
		autodoc.APIDoc{
			Method:      "POST",
			Name:        "Endpoint B",
			Description: "Post request with stuff",
			Request: autodoc.Values{
				"email": {"patrickoyarzun@gmail.com"},
			},
			Response: autodoc.JSON(map[string]interface{}{
				"Test": 2,
			}),
		},
		autodoc.APIDoc{
			Method:      "GET",
			Name:        "Endpoint C",
			Description: "Get all the A through json",
			Request: autodoc.Values{
				"email": {"patrickoyarzun@gmail.com"},
			},
			Response: autodoc.String("SUCCESS"),
		})

	http.ListenAndServe(":8080", doc)
}
