package html

import (
	"html/template"
	"net/http"
)

const errorTemplateContent = `
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>error | otf</title>
  <style>
  pre {
  	margin: auto;
	width: 60%;
	max-width: 72em;
	white-space: pre-wrap;
	border-style: solid;
	border-width: 1px;
	padding: 1em;
	}
  </style>
</head>
<body>
  <pre>{{ . }}</pre>
</body>
</html>
`

var errorTemplate = template.Must(template.New("error").Parse(errorTemplateContent))

func writeError(w http.ResponseWriter, err string, code int) {
	w.WriteHeader(code)

	errorTemplate.Execute(w, err)
}
