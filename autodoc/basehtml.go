package autodoc

var base_html = `<!DOCTYPE html>
<html>
<a id="top"></a>
<title>{{.Title}}</title>

<xmp theme="{{.Theme}}" style="display:none;">

# Contents
<hr />
Name | Method | URL
-----|--------|----
{{range .APIs}}<a href="#{{.Name | id}}">{{.Name}}</a> | {{.Method}} | {{.URL}}
{{end}}

# Details
<hr />
{{range .APIs}}

<a id="{{.Name | id}}"> </a>
## {{.Name}}
<a href="#">top</a>

{{.Description}}

    {{.Method}} {{.URL}}


{{if .Request}}
Request:

    {{.Request.Describe}}

{{end}}
{{if .Response}}
Response:

    {{.Response.Describe}}

{{end}}
{{end}}

</xmp>

<script src="http://strapdownjs.com/v/0.2/strapdown.js"></script>
</html>`
