package tmpl

import (
	"embed"
	"strings"
)

const (
	PHCustomImports = "/* CUSTOM IMPORTS */"
	PHCode          = "/* CODE */"
	PHFormalPara    = "/* FORMAL PARAMETERS */"
	PHActualPara    = "/* ACTUAL PARAMETERS */"
	PHPlugInfo      = "/* PLUGIN_INFO */"
	PHFunName       = "/* FUN_NAME */"
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

func GetTemplate(os, pType string) (string, error) {
	path := "templates"
	if pType == "fuzzTypes" {
		// path = filepath.Join(path, pType, pType+".gotmp")
		path = pathJoin(path, pType, pType+".gotmp")
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
