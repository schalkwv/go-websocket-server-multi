package handler

import (
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

func getTemplate() *template.Template {
	return template.Must(template.New("").Funcs(sprig.FuncMap()).Funcs(funcMap()).ParseGlob("templates/*.gohtml"))
}
