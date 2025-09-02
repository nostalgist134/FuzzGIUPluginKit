package goParser

import (
	"bytes"
	"fmt"
	"github.com/nostalgist134/FuzzGIUPluginKit/convention"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

// FindFunction 接收Go源码路径和函数名，返回对应的FuncDecl结构和参数元信息
func FindFunction(filePath, funcName string) (*convention.FuncDecl, []convention.ParaMeta, error) {
	// 创建文件集
	fset := token.NewFileSet()

	// 解析Go源文件，包括注释
	node, err := parser.ParseFile(fset, filePath, nil, parser.AllErrors|parser.ParseComments)
	if err != nil {
		return nil, nil, err
	}

	var funcDeclResult *convention.FuncDecl
	var paraMetas []convention.ParaMeta

	// 遍历AST查找函数声明
	ast.Inspect(node, func(n ast.Node) bool {
		// 如果已经找到结果，停止遍历
		if funcDeclResult != nil {
			return false
		}

		// 检查是否为函数声明
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		// 检查函数名是否匹配
		if funcDecl.Name.Name == funcName {
			// 提取函数参数和参数元信息
			params, metas := extractParamsWithComments(funcDecl.Type.Params, node.Comments)

			// 提取返回类型
			retType := extractReturnType(funcDecl.Type.Results)

			funcDeclResult = &convention.FuncDecl{
				Params:  params,
				RetType: retType,
			}
			paraMetas = metas
			return false // 找到后停止遍历
		}

		return true
	})

	if funcDeclResult == nil {
		return nil, nil, os.ErrNotExist // 函数未找到
	}

	return funcDeclResult, paraMetas, nil
}

// GetCode 读取Go源文件，返回除去package和import语句之外的所有内容
func GetCode(filePath string) (string, error) {
	// 创建文件集
	fset := token.NewFileSet()

	// 解析Go源文件
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return "", err
	}

	// 用于存储过滤后的声明
	var filteredDecls []ast.Decl

	// 遍历所有声明，过滤掉package和import语句
	for _, decl := range node.Decls {
		// 检查是否为package声明
		if genDecl, ok := decl.(*ast.GenDecl); ok {
			// 处理package声明
			if genDecl.Tok == token.PACKAGE {
				continue // 跳过package声明
			}
			// 处理import声明
			if genDecl.Tok == token.IMPORT {
				continue // 跳过import声明
			}
		}

		// 保留其他所有声明
		filteredDecls = append(filteredDecls, decl)
	}

	// 创建新的文件节点
	newFile := &ast.File{
		Name:     node.Name, // 保留包名（但移除了package语句）
		Decls:    filteredDecls,
		Comments: node.Comments, // 保留注释
	}

	// 将AST转换回代码
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, newFile); err != nil {
		return "", err
	}
	ret := buf.String()
	if strings.Index(ret, "package") == 0 {
		if len(strings.Split(ret, "\n")) > 1 {
			ret = ret[strings.Index(ret, "\n"):]
		} else {
			ret = ""
		}
	}
	return ret, nil
}

// GetImports 提取文件中的import列表
func GetImports(filename string, asSource ...bool) ([]string, error) {
	var src any = nil

	if len(asSource) > 0 && asSource[0] {
		// 传入的是源码内容字符串
		src = filename
		filename = "" // 设置为空让 parser.ParseFile 报错信息更通用
	} else {
		// 传入的是文件路径
		data, err := os.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		src = data
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, src, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}

	var imports []string
	for _, imp := range file.Imports {
		if imp.Name != nil {
			// 命名或特殊导入（_ 或 .）
			imports = append(imports, fmt.Sprintf("%s %s", imp.Name.Name, imp.Path.Value))
		} else {
			imports = append(imports, imp.Path.Value)
		}
	}

	return imports, nil
}

// 提取函数参数及相关注释信息
func extractParamsWithComments(fieldList *ast.FieldList, comments []*ast.CommentGroup) ([]convention.Param,
	[]convention.ParaMeta) {
	var params []convention.Param
	var paraMetas []convention.ParaMeta

	if fieldList == nil {
		return params, paraMetas
	}

	for _, field := range fieldList.List {
		typeStr := getTypeString(field.Type)

		// 获取参数名和类型的位置信息
		var namePos, typePos token.Pos
		if len(field.Names) > 0 {
			namePos = field.Names[0].Pos()
		} else {
			// 处理匿名参数
			namePos = field.Type.Pos() // 用类型位置作为参考
		}
		typePos = field.Type.Pos()

		// 查找参数名和类型之间的INFO注释
		infoComment := findInfoCommentBetween(namePos, typePos, comments)

		// 处理多个参数名共享同一类型的情况
		if len(field.Names) == 0 {
			// 没有参数名，可能是匿名参数或函数类型参数
			param := convention.Param{
				Name: "",
				Type: typeStr,
			}
			params = append(params, param)

			// 无论infoComment是否为空，都添加到元信息
			paraMetas = append(paraMetas, convention.ParaMeta{
				Param:    param,
				ParaInfo: infoComment, // 可能为空字符串
			})
		} else {
			for _, name := range field.Names {
				param := convention.Param{
					Name: name.Name,
					Type: typeStr,
				}
				params = append(params, param)

				// 无论infoComment是否为空，都添加到元信息
				paraMetas = append(paraMetas, convention.ParaMeta{
					Param:    param,
					ParaInfo: infoComment, // 可能为空字符串
				})
			}
		}
	}

	return params, paraMetas
}

// 在指定位置之间查找INFO注释
func findInfoCommentBetween(start, end token.Pos, comments []*ast.CommentGroup) string {
	for _, commentGroup := range comments {
		for _, comment := range commentGroup.List {
			// 检查注释是否在参数名和类型之间
			if comment.Pos() >= start && comment.End() <= end {
				// 检查是否是INFO注释
				text := comment.Text
				if strings.HasPrefix(text, "/*INFO:") && strings.HasSuffix(text, "*/") {
					// 提取INFO内容
					info := strings.TrimPrefix(text, "/*INFO:")
					info = strings.TrimSuffix(info, "*/")
					return strings.TrimSpace(info)
				}
			}
		}
	}

	return ""
}

// 提取返回类型
func extractReturnType(fieldList *ast.FieldList) string {
	if fieldList == nil || len(fieldList.List) == 0 {
		return ""
	}

	// 处理多个返回值的情况
	var types []string
	for _, field := range fieldList.List {
		types = append(types, getTypeString(field.Type))
	}

	return strings.Join(types, ", ")
}

// 将AST类型表达式转换为字符串
func getTypeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + getTypeString(t.X)
	case *ast.ArrayType:
		if t.Len != nil {
			// 固定长度数组
			return "[" + t.Len.(*ast.BasicLit).Value + "]" + getTypeString(t.Elt)
		}
		// 切片
		return "[]" + getTypeString(t.Elt)
	case *ast.MapType:
		return "map[" + getTypeString(t.Key) + "]" + getTypeString(t.Value)
	case *ast.ChanType:
		dir := ""
		switch t.Dir {
		case ast.SEND:
			dir = "<-"
		case ast.RECV:
			dir = "<-"
		}
		return "chan " + dir + getTypeString(t.Value)
	case *ast.FuncType:
		// 函数类型
		params, _ := extractParamsWithComments(t.Params, nil)
		var paramStrs []string
		for _, p := range params {
			paramStrs = append(paramStrs, p.Type)
		}
		retType := extractReturnType(t.Results)
		if retType != "" {
			return "func(" + strings.Join(paramStrs, ", ") + ") " + retType
		}
		return "func(" + strings.Join(paramStrs, ", ") + ")"
	case *ast.SelectorExpr:
		// 处理包名.类型的情况，如time.Time
		return getTypeString(t.X) + "." + t.Sel.Name
	}

	// 如果遇到未处理的类型，返回空字符串或占位符
	return ""
}
