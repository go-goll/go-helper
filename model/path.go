// Package model provides ...
package model

import (
	"bytes"
	"fmt"
	"io/fs"
	"os/exec"
	"path/filepath"
	"strings"
)

type fileInfo struct {
	name string // eg. example
	file string // eg. example.sql
	path string // eg. cmd/example/model/example/xx.sql

	dst string // eg. <dst>/example
	pkg string // eg. <dst>/example
}

func calculatePath(src, dst string) ([]fileInfo, error) {
	var files []fileInfo

	paths := strings.Split(src, ",")
	for _, v := range paths {
		err := filepath.WalkDir(v, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if name := d.Name(); filepath.Ext(name) == ".sql" {
				var folder string
				dir := filepath.Dir(path)
				if strings.HasPrefix(dir, dst) {
					folder, err = filepath.Rel(dst, dir)
					if err != nil {
						return err
					}
				}

				pkg := getPkgName(dst)
				files = append(files, fileInfo{
					name: strings.Replace(name, ".sql", "", 1),
					file: name,
					path: path,
					dst:  filepath.Join(dst, folder),
					pkg:  pkg,
				})
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return files, nil
}

// func cmdPkgPath(dst string) string {
// 	wd, err := os.Getwd()
// 	if err != nil {
// 		return ""
// 	}
// 	cmd := exec.Command("go", "env", "GOPATH")
// 	data, err := cmd.Output()
// 	if err != nil || len(data) == 0 {
// 		return ""
// 	}
// 	data = bytes.TrimSpace(data)
// 	path := filepath.Join(wd, dst)
// 	base := filepath.Join(string(data), "src")
// 	pkg, err := filepath.Rel(base, path)
// 	if err != nil {
// 		return ""
// 	}
// 	return pkg
// }

func getPkgName(dst string) string {
	cmd := exec.Command("go", "list", "-m")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		fmt.Println("执行 'go list -m' 失败:", err)
		return ""
	}

	moduleName := strings.TrimSpace(out.String())
	if moduleName == "" {
		fmt.Println("无法获取模块名，可能不在 Go 模块中")
		return ""
	}
	if !strings.HasPrefix(dst, "/") {
		return moduleName + "/" + dst
	}
	return moduleName + dst
}
