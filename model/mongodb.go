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

//go:embed template/internal_mgo.tmpl
var internalMgoTmpl string

//go:embed template/custom_mgo.tmpl
var customMgoTmpl string

//go:embed template/model_mgo.tmpl
var modelMgoTmpl string

type mongodbGenerator struct{}

func newMongoDBGenerator() (*mongodbGenerator, error) {
	// template
	err := ParseTemplate("internalMgoTmpl", internalMgoTmpl)
	if err != nil {
		return nil, err
	}
	err = ParseTemplate("customMgoTmpl", customMgoTmpl)
	if err != nil {
		return nil, err
	}
	err = ParseTemplate("modelMgoTmpl", modelMgoTmpl)
	if err != nil {
		return nil, err
	}
	return &mongodbGenerator{}, nil
}

func (mgo *mongodbGenerator) generateInternalFile(path string, params *commandParams) error {
	added := make(map[string]bool)
	buf := new(bytes.Buffer)
	buf.WriteString("	var idxs []mongo.IndexModel\n")
	for _, v := range params.Fields {
		for _, v := range v.indexs {
			var keys string
			for i, vv := range v.indexFields {
				if i != 0 {
					keys += ","
				}
				name := strcase.ToSnakeWithIgnore(vv.Name, "sha1")
				keys += fmt.Sprintf("{Key: \"%s\", Value: 1}", name)
			}
			if added[keys] {
				continue
			}
			switch {
			case v.uniqueIndex:
				buf.WriteString(fmt.Sprintf(`
	idxs = append(idxs, mongo.IndexModel{
		Keys: bson.D{%s},
		Options: options.Index().SetUnique(true),
	})
`, keys))

			case v.normalIndex:
				buf.WriteString(fmt.Sprintf(`
	idxs = append(idxs, mongo.IndexModel{
		Keys: bson.D{%s},
	})
`, keys))

			default:
				continue
			}
			added[keys] = true
		}
	}
	buf.WriteString("	mgo.Collection().Indexes().CreateMany(context.Background(), idxs)\n")
	params.MgoIndex = buf.String()

	buf.Reset()
	mgo.generateDeleteIndexDao(params, buf)
	mgo.generateUpdateIndexDao(params, buf)
	mgo.generateSelectIndexDao(params, buf)
	params.IndexGo = buf.String()

	buf.Reset()
	err := ExecuteTemplate(buf, "internalMgoTmpl", params)
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

func (mgo *mongodbGenerator) generateCustomFile(path string, params *commandParams) error {
	buf := &bytes.Buffer{}
	err := ExecuteTemplate(buf, "customMgoTmpl", params)
	if err != nil {
		return err
	}
	data, err := imports.Process("", buf.Bytes(), nil)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (mgo *mongodbGenerator) generateDeleteIndexDao(params *commandParams, buf *bytes.Buffer) {
	added := make(map[string]bool)
	for _, v := range params.Fields {
		for _, v := range v.indexs {
			var (
				key, input string
				filter     = "	filter := bson.M{\n"
			)
			for i, vv := range v.indexFields {
				key += vv.Name

				n := strcase.ToLowerCamel(vv.Name)
				if i == 0 {
					input = n + " " + vv.Type
				} else {
					input += ", " + n + " " + vv.Type
				}
				name := strcase.ToSnakeWithIgnore(vv.Name, "sha1")
				filter += fmt.Sprintf("		\"%s\": %s,\n", name, n)
			}
			filter += "	}\n"
			if (!v.uniqueIndex && !v.normalIndex) || added[key] {
				continue
			}
			added[key] = true

			funcName := fmt.Sprintf("Delete%sBy%s", params.TableName, key)
			// comments
			buf.WriteString(fmt.Sprintf("// %s delete object\n", funcName))
			// func
			buf.WriteString(fmt.Sprintf("func (d %sDao) %s", params.TableName, funcName))
			buf.WriteString(fmt.Sprintf("(%s)", input))
			buf.WriteString(" error {\n")
			// filter
			buf.WriteString(filter)
			// exp
			buf.WriteString("	_, err := d.Collection().DeleteOne(context.Background(), filter)\n")
			buf.WriteString("	return err\n")
			// quote
			buf.WriteString("}\n\n")
		}
	}
}

func (mgo *mongodbGenerator) generateUpdateIndexDao(params *commandParams, buf *bytes.Buffer) {
	added := make(map[string]bool)
	for _, v := range params.Fields {
		for _, v := range v.indexs {
			var (
				key, input string
				filter     = "	filter := bson.M{\n"
			)
			for i, vv := range v.indexFields {
				key += vv.Name
				n := strcase.ToLowerCamel(vv.Name)
				if i == 0 {
					input = n + " " + vv.Type
				} else {
					input += ", " + n + " " + vv.Type
				}
				name := strcase.ToSnakeWithIgnore(vv.Name, "sha1")
				filter += fmt.Sprintf("	\"%s\": %s,\n", name, n)
			}
			filter += "	}\n"
			if (!v.uniqueIndex && !v.normalIndex) || added[key] {
				continue
			}
			added[key] = true

			input += ", fields map[string]interface{}"

			funcName := fmt.Sprintf("Update%sBy%s", params.TableName, key)
			// comments
			buf.WriteString(fmt.Sprintf("// %s update object\n", funcName))
			// func
			buf.WriteString(fmt.Sprintf("func (d %sDao) %s", params.TableName, funcName))
			buf.WriteString(fmt.Sprintf("(%s)", input))
			buf.WriteString(" error {\n")
			// filter  & update
			filter += "	params := bson.M{}\n"
			filter += "	for k, v := range fields{\n"
			filter += "		params[k] = v\n"
			filter += "	}\n"
			filter += "	update := bson.M{\"$set\": params}\n"
			buf.WriteString(filter)
			// exp
			buf.WriteString("	_, err := d.Collection().UpdateOne(context.Background(), filter, update)\n")
			buf.WriteString("	return err\n")
			// quote
			buf.WriteString("}\n\n")
		}
	}
}

func (mgo *mongodbGenerator) generateSelectIndexDao(params *commandParams, buf *bytes.Buffer) {
	added := make(map[string]bool)
	for _, v := range params.Fields {
		for _, v := range v.indexs {
			var (
				key, input string
				filter     = "	filter := bson.M{\n"
			)
			for i, vv := range v.indexFields {
				key += vv.Name
				n := strcase.ToLowerCamel(vv.Name)
				if i == 0 {
					input = n + " " + vv.Type
				} else {
					input += ", " + n + " " + vv.Type
				}
				name := strcase.ToSnakeWithIgnore(vv.Name, "sha1")
				filter += fmt.Sprintf("		\"%s\": %s,\n", name, n)
			}
			filter += "	}\n"
			if (!v.uniqueIndex && !v.normalIndex) || added[key] {
				continue
			}
			added[key] = true

			funcName := fmt.Sprintf("Select%sBy%s", params.TableName, key)
			// comments
			buf.WriteString(fmt.Sprintf("// %s select object\n", funcName))
			// func
			buf.WriteString(fmt.Sprintf("func (d %sDao) %s", params.TableName, funcName))
			buf.WriteString(fmt.Sprintf("(%s)", input))
			buf.WriteString(fmt.Sprintf(" (*%sObj, error) {\n", params.TableName))
			// filter
			buf.WriteString(filter)
			// exp
			buf.WriteString(fmt.Sprintf("	obj := new(%sObj)\n", params.TableName))
			buf.WriteString("	err := d.Collection().FindOne(context.Background(), filter).Decode(obj)\n")
			buf.WriteString("	return obj, err\n")
			// quote
			buf.WriteString("}\n\n")
		}
	}
}

func (mgo *mongodbGenerator) generateModelFile(path string, list []*commandParams) error {
	buf := &bytes.Buffer{}
	err := ExecuteTemplate(buf, "modelMgoTmpl", list)
	if err != nil {
		return err
	}
	data, err := imports.Process("", buf.Bytes(), nil)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
