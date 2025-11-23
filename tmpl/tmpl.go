package tmpl

import (
	"embed"
	"io/fs"
	"strings"
)

const (
	PHCustomImports = "/* CUSTOM IMPORTS */"
	PHCode          = "/* CODE */"
	PHFormalPara    = "/* FORMAL PARAMETERS */"
	PHActualPara    = "/* ACTUAL PARAMETERS */"
	PHPlugInfo      = "/* PLUGIN_INFO */"
	PHFunName       = "/* FUN_NAME */"
	PHMinorFunName  = "/* MINOR_FUN_NAME */"
	PHModuleName    = "/* MODULE_NAME */"
)

//go:embed templates/**/*
var templates embed.FS

// pathJoin 连接各个路径（正斜杆）
func pathJoin(path string, path2 ...string) string {
	sb := strings.Builder{}
	sb.WriteString(strings.Replace(path, "\\", "/", -1))
	for _, p := range path2 {
		sb.WriteByte('/')
		sb.WriteString(p)
	}
	return sb.String()
}

type FileTmpl struct {
	Name    string
	Content []byte
}

// GetTemplatesDir 枚举某个子目录下所有模板文件（非递归）
func GetTemplatesDir(dir string) []FileTmpl {
	var tmpls []FileTmpl
	basePath := pathJoin("templates", dir)

	entries, err := fs.ReadDir(templates, basePath)
	if err != nil {
		return nil
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		fullPath := pathJoin(basePath, entry.Name())
		content, err := templates.ReadFile(fullPath)
		if err != nil {
			continue
		}
		tmpls = append(tmpls, FileTmpl{
			Name:    entry.Name(),
			Content: content,
		})
	}
	return tmpls
}

func GetTemplate(os, pType string) (string, error) {
	path := "templates"
	if strings.Index(pType, "fuzzTypes") == 0 {
		path = pathJoin(path, pType+".gotmp")
		// 傻逼embed库只支持正斜杆，不支持反斜杆，windows用filepath.Join反斜杠就打不开，妈了个逼的调半天才发现不是我的问题，吃大便去吧
		ft, err := templates.ReadFile(path)
		if err != nil {
			return "", err
		}
		return string(ft), err
	}
	if os == "windows" {
		path = pathJoin(path, "cgo")
	} else {
		path = pathJoin(path, "plugin")
	}
	fileName := strings.Title(pType) + ".gotmp"
	if pType != "pluginInfo" {
		fileName = "tmpl" + fileName
	}
	path = pathJoin(path, fileName)
	t, err := templates.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(t), nil
}

func Replace(src string, placeHolder string, to string) string {
	return strings.Replace(src, placeHolder, to, -1)
}
