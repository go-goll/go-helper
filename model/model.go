// Package model provides parse DDL to model, but only parse format without syntax
package model

import (
	_ "embed" // embed
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
)

// ModelCommand to generate model function
var ModelCommand = &cli.Command{
	Name:  "model",
	Usage: "to generate model CURD by .sql",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "force",
			Aliases: []string{"f"},
			Usage:   "Force rewrite model file: model/*.go",
		},
		&cli.StringFlag{
			Name:     "src",
			Usage:    "DDL file dir, eg. cmd/example/model",
			Required: true,
		},
		&cli.StringFlag{
			Name:  "dst",
			Usage: "DDL file generate dest dir, eg. cmd/ta-auth/model",
		},
		&cli.StringFlag{
			Name:  "driver",
			Usage: "DDL file generate for which DB, mongodb/postgres",
			Value: "postgres",
		},
	},
	Action: commandAction,
}

func commandAction(c *cli.Context) (err error) {
	src := c.String("src")
	fmt.Println("sql src: ", src)

	// new generator
	var generator fileGenerator
	switch c.String("driver") {
	case "postgres":
		generator, err = newPostgresGenerator()
	case "mongodb":
		generator, err = newMongoDBGenerator()
	}
	if err != nil {
		return err
	}
	dst := c.String("dst")
	if dst == "" {
		dst = src
	}
	files, err := calculatePath(src, dst)
	if err != nil {
		return err
	}
	// buf
	list := make([]*commandParams, len(files))
	for i, file := range files {
		// internal generate
		internalDir := filepath.Join(dst, "internal")
		_ = os.MkdirAll(internalDir, 0755)
		path := filepath.Join(internalDir, file.name+".go")
		var data []byte
		data, err = os.ReadFile(file.path)
		if err != nil {
			return err
		}
		var params *commandParams
		params, err = ddlAnalyzer(data)
		if err != nil {
			return err
		}
		// internal file
		err = generator.generateInternalFile(path, params)
		if err != nil {
			return err
		}

		// custom generate
		path = filepath.Join(file.dst, file.name+".go")
		_, params.PkgName = filepath.Split(file.dst)
		if _, err = os.Stat(path); os.IsNotExist(err) || c.Bool("f") {
			params.Import = filepath.Join(file.pkg, "internal")
			err = generator.generateCustomFile(path, params)
			if err != nil {
				return err
			}
			params.Import = ""
		}
		// model generate
		if strings.HasPrefix(file.path, file.dst) && file.path != filepath.Join(dst, file.file) {
			params.Import = filepath.Join(file.pkg, params.PkgName)
		}
		list[i] = params
	}
	path := filepath.Join(dst, "model.go")
	return generator.generateModelFile(path, list)
}

type commandParams struct {
	Import  string // model生成
	PkgName string // 文件名->pkg name

	TableName string
	Fields    []*field
	Primary   *primaryKey
	IndexGo   string // 索引语句

	MgoIndex string // mongodb index
}

type field struct {
	defaultVal string
	notNull    bool
	indexs     []index
	createdAt  bool
	updatedAt  bool

	Name    string // 字段名称
	Type    string // 数据类型
	Tag     string // tag
	Comment string // 备注
}

type index struct {
	uniqueIndex bool
	normalIndex bool
	indexFields []*field
}

type primaryKey struct {
	*field

	Autoincrement bool
	ShortID       bool
}

type fileGenerator interface {
	generateInternalFile(path string, params *commandParams) error
	generateCustomFile(path string, params *commandParams) error
	generateModelFile(path string, list []*commandParams) error
}
