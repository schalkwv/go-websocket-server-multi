package handler

import (
	"html/template"
)

// funcMap produces the function map.
//
// Use this to pass the functions into the template engine:
//
//	tpl := template.New("foo").Funcs(funcMap()))
func funcMap() template.FuncMap {
	return map[string]interface{}{
		"someFunction": func(id string) string {
			return "someFunction"
		},
	}
}
