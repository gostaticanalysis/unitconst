package unitconst

import (
	"go/token"
	"text/template"
)

type tmplData struct {
	Exprs map[token.Pos]string
	Types []*hashed
}

var srcTmpl = template.Must(template.New("a.go").Parse(`package a
{{range .Types}}
type {{.Name}} {{.Under}} // {{.Org}}
{{end}}

func f() {
{{- range $pos, $expr := .Exprs}}
	var v{{$pos}} = {{$expr}}
	_ = v{{$pos}}
{{end -}}
}
`))
