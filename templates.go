package templates

import (
	"html/template"
)

// TEMPLATES

var dirView = `<html>
  <body>
    {{template "dir" .Dir}}
  </body>
</html>`

var dir = `{{define "dir"}}
<div>
   <p>{{.Path}</p>
   {{range .}}
      <p>{{.Files.Name}}</p>
   {{end}}
</div>
{{end}}`

var fileView = `<html>
  <body>
    {{template "file" .File}}
  </body>
</html>`

var file = `{{define "file"}}
<div>
   <h1>{{.Name}}</h1>
   <h2>{{.Path}}</h2>
</div>
<p>{{.Content}}</p>`
