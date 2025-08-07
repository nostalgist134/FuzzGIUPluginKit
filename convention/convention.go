package convention

import (
	"FuzzGIUPluginKit/env"
	"FuzzGIUPluginKit/tmpl"
	"encoding/json"
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/plugin"
	"os"
	"strconv"
	"strings"
)

// GetPluginFunName 根据插件类型查找函数名
func GetPluginFunName(pluginType string) string {
	for i, t := range PluginTypes {
		if strings.ToLower(t) == strings.ToLower(pluginType) {
			return PluginFunNames[i]
		}
	}
	return ""
}

// GetPluginType 根据函数名查找插件类型
func GetPluginType(pluginFunName string) string {
	for i, n := range PluginFunNames {
		if n == pluginFunName {
			return PluginTypes[i]
		}
	}
	return ""
}

func GetFuncDecl(pluginType string) FuncDecl {
	if fd, ok := FuncDecls[pluginType]; ok {
		return fd
	}
	return FuncDecl{}
}

// CheckPluginFun 判断插件函数的函数声明是否符合规范
func CheckPluginFun(pluginType string, fd FuncDecl) (bool, string) {
	correctFd := GetFuncDecl(pluginType)
	funName := GetPluginFunName(pluginType)
	if correctFd.RetType != fd.RetType {
		return false, fmt.Sprintf("function %s return type incorrect, your return - %s, correct - %s",
			funName, fd.RetType, correctFd.RetType)
	}
	if len(fd.Params) < len(correctFd.Params) {
		return false, fmt.Sprintf("function %s argument not enough, at least %d arguments needed",
			funName, len(correctFd.Params))
	}
	for i, p := range correctFd.Params {
		if p.Name != fd.Params[i].Name || p.Type != fd.Params[i].Type {
			return false, fmt.Sprintf(`function %s param %d incorrect, your param: "%s %s". correct: "%s %s"`,
				funName, i, fd.Params[i].Name, fd.Params[i].Type, p.Name, p.Type)
		}
	}
	return true, ""
}

// genPluginFun 根据插件类型生成对应的插件函数
func genPluginFun(pluginType string) string {
	correctFd := GetFuncDecl(pluginType)
	funName := GetPluginFunName(pluginType)
	sb := strings.Builder{}
	// 参数列表
	for _, p := range correctFd.Params {
		sb.WriteString(p.Name + " " + p.Type)
		if pluginType != "reqSender" {
			sb.WriteString(", ")
		}
	}
	paraList := sb.String()
	if pluginType != "reqSender" {
		paraList += "/* CUSTOM ARGUMENTS HERE */"
	}
	// 返回值
	retVal := ""
	switch correctFd.RetType {
	case "string":
		retVal = "\"\""
	case "[]string":
		retVal = "[]string{}"
	default:
		retVal = fmt.Sprintf("&%s{}", strings.TrimPrefix(correctFd.RetType, "*"))
	}
	fn := fmt.Sprintf("func %s(%s) %s {\n	return %s\n}\n", funName, paraList, correctFd.RetType, retVal)
	return fn
}

// GenPlugInfoFun 生成PluginInfo函数
func GenPlugInfoFun(pName, pType, goVer, usageFile string, params []ParaMeta) string {
	usage, err := os.ReadFile(usageFile)
	if usageFile != "" && err != nil {
		fmt.Printf("read usage file failed - %v, set empty\n", err)
	}
	pi := PluginInfo{
		Name:      pName,
		Type:      pType,
		GoVersion: goVer,
		UsageInfo: string(usage),
		Params:    params,
	}
	j, _ := json.Marshal(pi)
	quoted := strconv.Quote(string(j))
	pFun, err := tmpl.GetTemplate(env.GlobEnv.OS, "pluginInfo")
	if err != nil {
		fmt.Printf("gen PluginInfo failed, reason: %v. skip\n", err)
	}
	pFun = tmpl.Replace(pFun, tmpl.PHPlugInfo, quoted)
	return pFun
}

// GetParamStrings 获取用于替换go模板文件中的参数占位符的字符串
func GetParamStrings(os, pluginType string, params []Param) (string, string) {
	correctFd := GetFuncDecl(pluginType)
	if len(correctFd.Params) > len(params) {
		return "", ""
	}
	formalParams := strings.Builder{}
	actualParams := strings.Builder{}
	// 非windows（使用plugin库）
	if os != "windows" {
		/*
			1.由于plugin库的特性，调用函数前必须断言为一个固定的函数类型，但是要支持用户自定义参数，因此形参只能写成...any
			2.写好的插件函数其中的参数类型就已经固定下来了，但是PluginWrapper的参数列表是any类型，因此用户自定义实参列表需要按顺序类型断言
		*/
		formalParams.WriteString("args ...any")
		//从固定参数之后开始动态补全用户自定义参数（下标也必须对应）
		for i := len(correctFd.Params); i < len(params); i++ {
			actualParams.WriteString(fmt.Sprintf("args[%d].(%s),", i, params[i].Type))
		}
	} else {
		for i := len(correctFd.Params); i < len(params); i++ {
			formalParams.WriteString(fmt.Sprintf("%s %s,", params[i].Name, params[i].Type))
			actualParams.WriteString(params[i].Name + ",")
		}
	}
	return formalParams.String(), actualParams.String()
}

func GetPreDefinedArgs(pType string) []any {
	switch pType {
	case PluginTypes[IndPTypeReqSender]:
		return []any{new(fuzzTypes.SendMeta)}
	case PluginTypes[IndPTypePreproc]:
		return []any{new(fuzzTypes.Fuzz)}
	case PluginTypes[IndPTypeReact]:
		return []any{new(fuzzTypes.Req), new(fuzzTypes.Resp)}
	}
	return nil
}

// BuildFd 根据插件实际信息返回一个FuncDecl结构
func BuildFd(inf *PluginInfo) FuncDecl {
	params := make([]Param, 0)
	fd := FuncDecl{}
	for _, pm := range inf.Params {
		params = append(params, pm.Param)
	}
	fd.RetType = GetFuncDecl(inf.Type).RetType
	fd.Params = params
	return fd
}

// 根据函数声明确定是否需要import语句
func needImport(fd FuncDecl) string {
	imp := "\nimport \"/* MODULE_NAME *//components/fuzzTypes\"\n"
	if strings.Index(fd.RetType, "fuzzTypes") != -1 {
		return imp
	} else {
		for _, p := range fd.Params {
			if strings.Index(p.Type, "fuzzTypes") != -1 {
				return imp
			}
		}
	}
	return ""
}

// GenCodeByType 根据插件类型生成一个完整的可通过编译的代码骨架
func GenCodeByType(pluginType string) string {
	fn := genPluginFun(pluginType)
	imp := needImport(GetFuncDecl(pluginType))
	return fmt.Sprintf("package main\n%s\n%s\n", imp, fn)
}

func GetStruct(structType string) any {
	switch structType {
	case "*fuzzTypes.Fuzz":
		return &fuzzTypes.Fuzz{}
	case "*fuzzTypes.Req":
		return &fuzzTypes.Req{}
	case "*fuzzTypes.Resp":
		return &fuzzTypes.Resp{}
	case "*fuzzTypes.SendMeta":
		return &fuzzTypes.SendMeta{}
	case "*fuzzTypes.Reaction":
		return &fuzzTypes.Reaction{}
	default:
		return nil
	}
}

// GetRetPtr 取得指向插件返回值的指针
func GetRetPtr(retType string) any {
	switch retType {
	case "string":
		s := ""
		return &s
	case "[]string":
		sSlice := make([]string, 0)
		return &sSlice
	default:
		return GetStruct(retType)
	}
}

// GetFullStruct 获取一个填写完整的结构体指针
func GetFullStruct(structType string) any {
	switch structType {
	case "*fuzzTypes.Fuzz":
		return fullFuzz
	case "*fuzzTypes.Req":
		return fullReq
	case "*fuzzTypes.Resp":
		return fullResp
	case "*fuzzTypes.SendMeta":
		return fullSendMeta
	case "*fuzzTypes.Reaction":
		return fullReaction
	default:
		return nil
	}
}

// GetPluginPathByPType 根据插件类型获取插件所在相对目录
func GetPluginPathByPType(pType string) string {
	ret := plugin.BaseDir
	switch pType {
	case PluginTypes[IndPTypePlGen]:
		ret += plugin.RelPathPlGen
	case PluginTypes[IndPTypePreproc]:
		ret += plugin.RelPathPreprocessor
	case PluginTypes[IndPTypeReqSender]:
		ret += plugin.RelPathReqSender
	case PluginTypes[IndPTypePlProc]:
		ret += plugin.RelPathPlProc
	case PluginTypes[IndPTypeReact]:
		ret += plugin.RelPathReactor
	default:
		ret = ""
	}
	return ret
}
