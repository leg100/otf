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

// Error sends an appropriate error response to an http request. If the request
// was to carry out an operation, i.e. a POST action, then a flash message is
// set and the user is redirected to the last page. Otherwise it's assumed the
// request was a normal page navigation request, i.e. a GET action, and an error
// notice is rendered with the given http code.
func Error(r *http.Request, w http.ResponseWriter, err string, code int) {
	if r.Method == "POST" && r.Referer() != "" {
		FlashError(w, err)
		http.Redirect(w, r, r.Referer(), http.StatusFound)
	} else {
		w.WriteHeader(code)
		errorTemplate.Execute(w, err)
	}
}
