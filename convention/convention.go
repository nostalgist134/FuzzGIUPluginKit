package convention

import (
	"encoding/json"
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIUPluginKit/env"
	"github.com/nostalgist134/FuzzGIUPluginKit/tmpl"
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
	for t, fd := range FuncDecls {
		if strings.ToLower(t) == strings.ToLower(pluginType) {
			return fd
		}
	}
	return FuncDecl{}
}

func compareFuncDecl(fd FuncDecl, correctFd FuncDecl) (equal bool, reason string) {
	if len(fd.Params) < len(correctFd.Params) {
		equal = false
		reason = fmt.Sprintf("argument count not enough, at least %d arguments needed", len(correctFd.Params))
		return
	}
	for i, p := range correctFd.Params {
		if p.Name != fd.Params[i].Name || p.Type != fd.Params[i].Type {
			equal, reason = false, fmt.Sprintf(`param %d incorrect, given: "%s %s", wanted: "%s %s"`,
				i, fd.Params[i].Name, fd.Params[i].Type, p.Name, p.Type)
			return
		}
	}
	if fd.RetType != correctFd.RetType {
		equal, reason = false, fmt.Sprintf(`return type incorrect, given: "%s", wanted: "%s"`, fd.RetType,
			correctFd.RetType)
	}
	equal = true
	return
}

// CheckPluginFun 判断插件函数的函数声明是否符合规范
func CheckPluginFun(pluginType string, fd FuncDecl) (bool, string) {
	correctFd := GetFuncDecl(pluginType)
	return compareFuncDecl(fd, correctFd)
}

func CheckPluginMinorFunc(pType string, fd FuncDecl) (bool, string) {
	if pType != PluginTypes[IndPTypeIterator] {
		return true, ""
	}
	correctFd := FuncDecls[PluginMinorFun[IndPTypeIteratorMinor]]
	return compareFuncDecl(fd, correctFd)
}

// genPluginFun 根据插件类型生成对应的插件函数
func genPluginFun(pluginType string) string {
	correctFd := GetFuncDecl(pluginType)
	funName := GetPluginFunName(pluginType)
	sb := strings.Builder{}
	// 参数列表
	for _, p := range correctFd.Params {
		sb.WriteString(p.Name + " " + p.Type)
		if strings.ToLower(pluginType) != strings.ToLower(PluginTypes[IndPTypeRequester]) {
			sb.WriteString(", ")
		}
	}
	paraList := sb.String()
	if strings.ToLower(pluginType) != strings.ToLower(PluginTypes[IndPTypeRequester]) {
		paraList += "/* CUSTOM ARGUMENTS HERE */"
	}
	// 返回值
	retVal := ""
	switch correctFd.RetType {
	case "string":
		retVal = "\"\""
	case "[]string":
		retVal = "[]string{}"
	case "[]int":
		retVal = "[]int{}"
	default:
		retVal = fmt.Sprintf("&%s{}", strings.TrimPrefix(correctFd.RetType, "*"))
	}
	fn := fmt.Sprintf("func %s(%s) %s {\n	// IMPLEMENT YOUR CODE HERE\n"+
		"	return %s\n}\n", funName, paraList, correctFd.RetType, retVal)
	return fn
}

// GenPlugInfoFun 生成PluginInfo函数
func GenPlugInfoFun(pName, pType, goVer, usageFile string, params []ParaMeta) string {
	var usage string
	if usageFile != "" {
		b, err := os.ReadFile(usageFile)
		if err != nil {
			fmt.Printf("read usage file failed - %v, set empty\n", err)
		} else {
			usage = string(b)
		}
	}
	pi := PluginInfo{
		Name:      pName,
		Type:      pType,
		GoVersion: goVer,
		UsageInfo: usage,
		Params:    params,
	}
	j, _ := json.Marshal(pi)
	quoted := strconv.Quote(string(j))
	pFun, err := tmpl.GetTemplate(env.GlobEnv.OS, "pluginInfo")
	if err != nil {
		fmt.Printf("warning: gen PluginInfo failed: %v\n", err)
	}
	pFun = tmpl.Replace(pFun, tmpl.PHPlugInfo, quoted)
	return pFun
}

// GetParamStrings 获取用于替换go模板文件中的参数占位符的字符串
func GetParamStrings(os, pluginType string, params []Param) (formal string, actual string) {
	correctFd := GetFuncDecl(pluginType)
	if len(correctFd.Params) > len(params) {
		return
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
	formal = formalParams.String()
	actual = actualParams.String()
	return
}

// GetContextArgs 根据插件类型返回预留参数
func GetContextArgs(pType string) []any {
	switch pType {
	case PluginTypes[IndPTypeRequester]:
		return []any{new(fuzzTypes.RequestCtx)}
	case PluginTypes[IndPTypePreproc]:
		return []any{new(fuzzTypes.Fuzz)}
	case PluginTypes[IndPTypeReact]:
		return []any{new(fuzzTypes.Req), new(fuzzTypes.Resp)}
	case PluginTypes[IndPTypeIterator]:
		return []any{make([]int, 10), 1}
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

// GenCodePType 根据插件类型生成一个完整的可通过编译的代码骨架
func GenCodePType(pluginType string) string {
	fn := genPluginFun(pluginType)
	imp := needImport(GetFuncDecl(pluginType))
	return fmt.Sprintf("package main\n%s\n%s\n", imp, fn)
}

// DefMinorFun 次要函数，目前只有iterator插件的IterLen使用这个函数生成
func DefMinorFun(pluginType string) string {
	if pluginType != PluginTypes[IndPTypeIterator] {
		return ""
	}
	return `func IterLen(lengths []int, /* FORMAL PARAMETERS */) int {
	return -1
}
`
}

func GetStruct(structType string) any {
	switch structType {
	case "*fuzzTypes.Fuzz":
		return &fuzzTypes.Fuzz{}
	case "*fuzzTypes.Req":
		return &fuzzTypes.Req{}
	case "*fuzzTypes.Resp":
		return &fuzzTypes.Resp{}
	case "*fuzzTypes.RequestCtx":
		return &fuzzTypes.RequestCtx{}
	case "*fuzzTypes.Reaction":
		return &fuzzTypes.Reaction{}
	case "[]int":
		return []int{}
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
	case "*fuzzTypes.RequestCtx":
		return fullRequestCtx
	case "*fuzzTypes.Reaction":
		return fullReaction
	default:
		return nil
	}
}
