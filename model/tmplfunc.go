// Package helper provides ...
package model

import (
	"io"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
)

func init() {
	strcase.ConfigureAcronym("id", "ID")
	strcase.ConfigureAcronym("ip", "IP")

	templateIns = template.New("zero").Funcs(template.FuncMap{
		"toLowerCamel": strcase.ToLowerCamel,
		"toSnake":      strcase.ToSnake,
		"stringsJoin":  strings.Join,
	})
}

// templateIns with function
var templateIns *template.Template

// ParseTemplate parse text
func ParseTemplate(name string, text string) error {
	t := templateIns.New(name)
	_, err := t.Parse(text)
	return err
}

// ExecuteTemplate execute template
func ExecuteTemplate(wr io.Writer, name string, params interface{}) error {
	return templateIns.ExecuteTemplate(wr, name, params)
}
