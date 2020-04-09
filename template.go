package unitconst

import (
	"go/token"
	"text/template"
)

type tmplData struct {
	Exprs map[token.Pos]string
	Pkgs  []string
	Types []string
}

var srcTmpl = template.Must(template.New("a.go").Parse(`package a
import (
{{- range .Pkgs}}
	"{{.}}"
{{end -}}
)

// dummy
var (
{{- range .Types}}
	_ {{.}}
{{end -}}
)

func f() {
{{- range $pos, $expr := .Exprs}}
	var v{{$pos}} = {{$expr}}
	_ = v{{$pos}}
{{end -}}
}
`))
