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
  {{ if .DevMode }}<script src="http://localhost:35729/livereload.js"></script>{{ end }}
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
  <pre>{{ .Error }}</pre>
</body>
</html>
`

var errorTemplate = template.Must(template.New("error").Parse(errorTemplateContent))

func Error(w http.ResponseWriter, err string, code int, devMode bool) {
	w.WriteHeader(code)

	errorTemplate.Execute(w, struct {
		Error   string
		DevMode bool
	}{
		Error:   err,
		DevMode: devMode,
	})
}
