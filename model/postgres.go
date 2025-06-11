// Package model provides ...
package model

import (
	"bytes"
	_ "embed" // embed
	"fmt"
	"os"

	"github.com/iancoleman/strcase"
	"golang.org/x/tools/imports"
)

//go:embed template/internal_pg.tmpl
var internalPGTmpl string

//go:embed template/custom_pg.tmpl
var customPGTmpl string

//go:embed template/model_pg.tmpl
var modelPGTmpl string

type postgresGenerator struct{}

func newPostgresGenerator() (*postgresGenerator, error) {
	// template
	err := ParseTemplate("internalPGTmpl", internalPGTmpl)
	if err != nil {
		return nil, err
	}
	err = ParseTemplate("customPGTmpl", customPGTmpl)
	if err != nil {
		return nil, err
	}
	err = ParseTemplate("modelPGTmpl", modelPGTmpl)
	if err != nil {
		return nil, err
	}
	return &postgresGenerator{}, nil
}

func (pg *postgresGenerator) generateInternalFile(path string, params *commandParams) error {
	buf := new(bytes.Buffer)
	pg.generateDeleteIndexDao(params, buf)
	pg.generateUpdateIndexDao(params, buf)
	pg.generateSelectIndexDao(params, buf)
	params.IndexGo = buf.String()

	buf.Reset()
	err := ExecuteTemplate(buf, "internalPGTmpl", params)
	if err != nil {
		return err
	}
	// imports
	data, err := imports.Process("", buf.Bytes(), nil)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (pg *postgresGenerator) generateCustomFile(path string, params *commandParams) error {
	buf := &bytes.Buffer{}
	err := ExecuteTemplate(buf, "customPGTmpl", params)
	if err != nil {
		return err
	}
	data, err := imports.Process("", buf.Bytes(), nil)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (pg *postgresGenerator) generateDeleteIndexDao(params *commandParams, buf *bytes.Buffer) {
	added := make(map[string]bool)

	// 收集所有唯一索引
	uniqueIndexes := make(map[string]index)
	for _, v := range params.Fields {
		for _, idx := range v.indexs {
			if idx.uniqueIndex && idx.indexName != "" {
				uniqueIndexes[idx.indexName] = idx
			}
		}
	}

	// 为每个唯一索引生成删除方法
	for indexName, idx := range uniqueIndexes {
		var (
			key, input string
			w, q       string
		)
		for i, vv := range idx.indexFields {
			key += vv.Name

			n := strcase.ToLowerCamel(vv.Name)
			name := strcase.ToSnakeWithIgnore(vv.Name, "sha1")
			if i == 0 {
				input = n + " " + vv.Type
				w = name + "=?"
				q = n
			} else {
				input += ", " + n + " " + vv.Type
				w += " AND " + name + "=?"
				q += ", " + n
			}
		}

		if added[key] {
			continue
		}
		added[key] = true

		funcName := fmt.Sprintf("Delete%sBy%s", params.TableName, key)
		// comments
		buf.WriteString(fmt.Sprintf("// %s delete object by unique index %s\n", funcName, indexName))
		// func
		buf.WriteString(fmt.Sprintf("func (d %sDao) %s", params.TableName, funcName))
		buf.WriteString(fmt.Sprintf("(%s)", input))
		buf.WriteString(" error {\n")
		// exp
		buf.WriteString(fmt.Sprintf(`	return d.DB.Where("%s", %s)`, w, q))
		buf.WriteString(fmt.Sprintf(".Delete(%sObj{}).Error\n", params.TableName))
		// quote
		buf.WriteString("}\n\n")
	}
}

func (pg *postgresGenerator) generateUpdateIndexDao(params *commandParams, buf *bytes.Buffer) {
	added := make(map[string]bool)

	// 收集所有唯一索引
	uniqueIndexes := make(map[string]index)
	for _, v := range params.Fields {
		for _, idx := range v.indexs {
			if idx.uniqueIndex && idx.indexName != "" {
				uniqueIndexes[idx.indexName] = idx
			}
		}
	}

	// 为每个唯一索引生成更新方法
	for indexName, idx := range uniqueIndexes {
		var (
			key, input string
			w, q       string
		)
		for i, vv := range idx.indexFields {
			key += vv.Name
			n := strcase.ToLowerCamel(vv.Name)
			name := strcase.ToSnakeWithIgnore(vv.Name, "sha1")
			if i == 0 {
				input = n + " " + vv.Type
				w = name + "=?"
				q = n
			} else {
				input += ", " + n + " " + vv.Type
				w += " AND " + name + "=?"
				q += ", " + n
			}
		}

		if added[key] {
			continue
		}
		added[key] = true

		input += ", fields map[string]interface{}"

		funcName := fmt.Sprintf("Update%sBy%s", params.TableName, key)
		// comments
		buf.WriteString(fmt.Sprintf("// %s update object by unique index %s\n", funcName, indexName))
		// func
		buf.WriteString(fmt.Sprintf("func (d %sDao) %s", params.TableName, funcName))
		buf.WriteString(fmt.Sprintf("(%s)", input))
		buf.WriteString(" error {\n")
		// exp
		buf.WriteString(fmt.Sprintf(`	return d.DB.Model(%sObj{}).Where("%s", %s)`,
			params.TableName, w, q))
		buf.WriteString(".Updates(fields).Error\n")
		// quote
		buf.WriteString("}\n\n")
	}
}

func (pg *postgresGenerator) generateSelectIndexDao(params *commandParams, buf *bytes.Buffer) {
	added := make(map[string]bool)

	// 收集所有唯一索引
	uniqueIndexes := make(map[string]index)
	for _, v := range params.Fields {
		for _, idx := range v.indexs {
			if idx.uniqueIndex && idx.indexName != "" {
				uniqueIndexes[idx.indexName] = idx
			}
		}
	}

	// 为每个唯一索引生成查询方法
	for indexName, idx := range uniqueIndexes {
		var (
			key, input string
			w, q       string
		)
		for i, vv := range idx.indexFields {
			key += vv.Name
			n := strcase.ToLowerCamel(vv.Name)
			name := strcase.ToSnakeWithIgnore(vv.Name, "sha1")
			if i == 0 {
				input = n + " " + vv.Type
				w = name + "=?"
				q = n
			} else {
				input += ", " + n + " " + vv.Type
				w += " AND " + name + "=?"
				q += ", " + n
			}
		}

		if added[key] {
			continue
		}
		added[key] = true

		funcName := fmt.Sprintf("Select%sBy%s", params.TableName, key)
		// comments
		buf.WriteString(fmt.Sprintf("// %s select object by unique index %s\n", funcName, indexName))
		// func
		buf.WriteString(fmt.Sprintf("func (d %sDao) %s", params.TableName, funcName))
		buf.WriteString(fmt.Sprintf("(%s)", input))
		buf.WriteString(fmt.Sprintf(" (*%sObj, error) {\n", params.TableName))
		// exp
		buf.WriteString(fmt.Sprintf("	obj := new(%sObj)\n", params.TableName))
		buf.WriteString(fmt.Sprintf(`	err := d.DB.Where("%s", %s)`, w, q))
		buf.WriteString(".First(obj).Error\n")
		buf.WriteString("	return obj, err\n")
		// quote
		buf.WriteString("}\n\n")
	}
}

func (pg *postgresGenerator) generateModelFile(path string, list []*commandParams) error {
	buf := &bytes.Buffer{}
	err := ExecuteTemplate(buf, "modelPGTmpl", list)
	if err != nil {
		return err
	}
	data, err := imports.Process("", buf.Bytes(), nil)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
