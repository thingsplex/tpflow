package http

import (
	"bytes"
	"html/template"
	"net/http"
)

func (conn *Connector) index(w http.ResponseWriter, r *http.Request) {
	const indexTemplate = `
    <html lang="en">
    <head> <title> TPflow http/ws endpoints </title> </head>
    <body>
    <p>List of endpoints</p>
    {{ range $index, $element := . }} 
      {{ if $element.IsWs }}
        <p> Websocket  : <a href="/flow/{{ $index }}/ws"> /flow/{{ $index }}/ws </a> - {{ $element.Name }} </p> 
      {{ else }} 
		<p> Http : <a href="/flow/{{ $index }}/rest"> /flow/{{ $index }}/rest </a> - {{ $element.Name }} </p> 
      {{ end }} 
    {{ end }}
    </body>
    </html>
    `
	t, err := template.New("foo").Parse(indexTemplate)
	if err != nil {
		return
	}
	var out bytes.Buffer
	err = t.Execute(&out, conn.flowStreamRegistry)
	w.Write(out.Bytes())
}
