// Package model provides ...
package model

import (
	"bytes"
	"errors"
	"regexp"
	"strings"

	"github.com/iancoleman/strcase"
)

var sqlTypeToGo = map[string]string{
	// int
	"SERIAL":   "int",
	"INTEGER":  "int",
	"SMALLINT": "int32",
	// int64
	"BIGINT": "int64",
	// bool
	"BOOLEAN": "bool",
	// string
	"TEXT":    "string",
	"VARCHAR": "string",
	// bytes
	"BYTEA": "[]byte",
	// time
	"TIMESTAMP": "time.Time",
	// json
	"JSON":  "json.RawMessage",
	"JSONB": "json.RawMessage",
	// special type
	"TEXT[]":    "db.StringArray",
	"INTEGER[]": "db.Int64Array",
}

func ddlAnalyzer(raw []byte) (params *commandParams, err error) {
	params = &commandParams{}

	data := bytesSplitAndTrimSpace(raw, []byte(";"))
	for _, v := range data {
		ddl := string(v)
		if ddl == "" {
			return params, nil
		}
		m := &marker{
			linesDDL:  strSplitAndTrimSpace(ddl, "\n"),
			lineIndex: -1,
		}

		line := m.nextLine()
		fields := m.Fields(line)
		switch strings.ToUpper(fields[0]) {
		case "CREATE":
			err = parseCreate(m, params)
		case "COMMENT":
			err = parseComment(m, params)
		case "ALTER":
			err = parseAlter(m, params)
		case "DROP":
			err = parseDrop(m, params)
		default:
			return nil, errors.New("unknown marker action: " + fields[0])
		}
		if err != nil {
			return nil, err
		}
	}
	// params
	return params, nil
}

func parseCreate(m *marker, params *commandParams) error {
	line := m.currentLine()
	line = strings.Replace(line, "IF NOT EXISTS", "", 1)
	fields := m.Fields(line)

	switch strings.ToUpper(fields[1]) {
	case "TABLE":
		params.TableName = strcase.ToCamel(fields[2])
		err := parseFields(m, params)
		if err != nil {
			return err
		}

		// rebuild tag
		rebuildTag := func(f *field) {
			// other field
			f.Tag = "column:" + f.Name
			if f.defaultVal != "" {
				f.Tag += ";default:" + f.defaultVal
			}
			if f.notNull {
				f.Tag += ";not null"
			}
			// 计算idx
			for _, idx := range f.indexs {
				var idxStr string
				for i, v := range idx.indexFields {
					if i == 0 {
						idxStr = "idx_" + fields[2]
					}
					idxStr += "_" + strcase.ToSnakeWithIgnore(v.Name, "sha1")
				}
				if idx.normalIndex {
					f.Tag += ";index:" + idxStr
				}
				if idx.uniqueIndex {
					f.Tag += ";uniqueIndex:" + idxStr
				}
			}
			if f.createdAt {
				f.Tag += ";autoCreateTime"
			}
			if f.updatedAt {
				f.Tag += ";autoUpdateTime"
			}
			if f.Type == "[]string" || f.Type == "[]int" {
				f.Tag += ";serializer:json"
			}
			// change name
			f.Name = strcase.ToCamel(f.Name)
		}
		// primary key
		if params.Primary != nil {
			rebuildTag(params.Primary.field)

			params.Primary.Tag += ";primaryKey"
			if params.Primary.Autoincrement {
				params.Primary.Tag += ";autoIncrement"
			}
		}
		// change content
		for _, v := range params.Fields {
			vt := strings.ToUpper(v.Type)
			// replace type
			ok := false
			for key, val := range sqlTypeToGo {
				// 判断filed是否位deleted
				if v.Name == "deleted_at" {
					v.Type = "gorm.DeletedAt"
					ok = true

					break
				}
				// 映射数据类型
				if vt == key {
					v.Type = val
					ok = true

					break
				}
			}
			if !ok {
				return errors.New("unsupported pg type to go:" + v.Type)
			}
			// rebuild tag
			if v.Tag == "" {
				rebuildTag(v)
			}
		}
		return nil
	}
	return errors.New("unsupported operation: CREATE " + fields[1])
}

func parseComment(m *marker, params *commandParams) error {
	line := m.currentLine()

	fields := m.Fields(line)
	if len(fields) < 6 {
		return errors.New("invalid ddl: " + line)
	}
	column := fields[3]
	// NOTE 统一sql文件只能操作同一个表，这里不判断表
	if sli := strSplitAndTrimSpace(fields[3], "."); len(sli) > 1 {
		column = sli[1]
	}
	comment := strings.Join(fields[5:], " ")

	idx := foundFiled(params.Fields, strcase.ToCamel(column))
	if idx > -1 {
		// params.Fields[idx].Comment = strings.Trim(comment, "'")
		params.Fields[idx].Tag += ";comment:" + strings.Trim(comment, "'")
	}
	return nil
}

func parseAlter(m *marker, params *commandParams) error {
	// TODO
	return nil
}

func parseDrop(m *marker, params *commandParams) error {
	// TODO
	return nil
}

var regexpIndex = regexp.MustCompile(`\((.+?)\)`)

func parseFields(m *marker, params *commandParams) error {
	for line := m.nextLine(); line != ""; line = m.nextLine() {
		fields := m.Fields(line)

		switch fields[0] {
		case "PRIMARY":
			if fields[1] != "KEY" {
				return errors.New("invalid ddl:" + line)
			}
			sub := regexpIndex.FindStringSubmatch(line)
			for i, v := range sub {
				if i == 0 {
					continue
				}
				// private key
				idx := foundFiled(params.Fields, v)
				if idx >= 0 {
					params.Primary = &primaryKey{field: params.Fields[idx]}
					if params.Primary.Type == "SERIAL" {
						params.Primary.Autoincrement = true
					} else if params.Primary.Name == "id" {
						params.Primary.ShortID = true
					}
				}
			}

		case "UNIQUE":
			sub := regexpIndex.FindStringSubmatch(line)
			for i, v := range sub {
				if i == 0 {
					continue
				}
				// unique
				arr := strSplitAndTrimSpace(v, ",")
				arrField := make([]*field, len(arr))
				for i, vv := range arr {
					idx := foundFiled(params.Fields, vv)
					if idx >= 0 {
						arrField[i] = params.Fields[idx]

						params.Fields[idx].indexs = append(params.Fields[idx].indexs, index{
							uniqueIndex: true,
							indexFields: arrField,
						})
					}
				}
			}
		case "INDEX":
			sub := regexpIndex.FindStringSubmatch(line)
			for i, v := range sub {
				if i == 0 {
					continue
				}
				// unique
				arr := strSplitAndTrimSpace(v, ",")
				arrField := make([]*field, len(arr))
				for i, vv := range arr {
					idx := foundFiled(params.Fields, vv)
					if idx >= 0 {
						arrField[i] = params.Fields[idx]

						params.Fields[idx].indexs = append(params.Fields[idx].indexs, index{
							normalIndex: true,
							indexFields: arrField,
						})
					}
				}
			}
		default: // 普通字段
			field := &field{
				Name: fields[0],
				Type: fields[1],
			}
			if i := foundFiled(fields, "DEFAULT"); i > 0 {
				field.defaultVal = fields[i+1]
			}
			if i := foundFiled(fields, "NOT"); i > 0 {
				if fields[i+1] != "NULL" {
					return errors.New("invalid ddl:" + line)
				}
				field.notNull = true
			}
			// 判断是否是create_at
			if field.Name == "created_at" {
				field.createdAt = true
			}
			if field.Name == "updated_at" {
				field.updatedAt = true
			}

			params.Fields = append(params.Fields, field)
		}
	}
	return nil
}

func foundFiled(fields interface{}, name string) int {
	switch fs := fields.(type) {
	case []string:
		for i, v := range fs {
			if v != name {
				continue
			}
			return i
		}
	case []*field:
		for i, v := range fs {
			if v.Name != name {
				continue
			}
			return i
		}
	}

	return -1
}

func strSplitAndTrimSpace(str string, sep string) []string {
	arr := strings.Split(str, sep)
	for i, v := range arr {
		arr[i] = strings.TrimSpace(v)
	}
	return arr
}

func bytesSplitAndTrimSpace(raw []byte, sep []byte) [][]byte {
	arr := bytes.Split(raw, sep)
	for i, v := range arr {
		arr[i] = bytes.TrimSpace(v)
	}
	return arr
}

type marker struct {
	linesDDL  []string
	lineIndex int
}

func (m *marker) nextLine() string {
	var currentLine string

	length := len(m.linesDDL)
	for m.lineIndex+1 < length {
		m.lineIndex++

		line := strings.TrimSpace(m.linesDDL[m.lineIndex])
		if line == "" || line == "(" {
			continue
		}
		if strings.HasPrefix(line, "--") {
			continue
		}
		if strings.HasPrefix(line, ")") {
			continue
		}
		currentLine = line

		break
	}
	currentLine = strings.TrimSuffix(currentLine, ",")
	return strings.TrimSpace(currentLine)
}

func (m *marker) currentLine() string {
	return m.linesDDL[m.lineIndex]
}

func (m *marker) Fields(line string) []string {
	fields := strings.Fields(line)

	for i, v := range fields {
		fields[i] = strings.TrimSpace(strings.Trim(v, "\""))
	}
	return fields
}
